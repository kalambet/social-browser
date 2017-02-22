package social

import (
	"appengine"
	"appengine/taskqueue"
	"encoding/json"
	"log"
	"net/url"
)

const (
	FetchNewUser  = "FetchNewUser"
	FetchOldUser  = "FetchOldUser"
	FetchUserPage = "FetchUserPage"
	PullQueueName = "actions-queue"
	PushQueueName = "workers-queue"
)

type Payload struct {
	Username string
	Page     int
}

// AddTaskInQueue creates PULL task in Google Task Queues
// tagging it by type of operation it meant to perform
func AddTaskInQueue(ctx *appengine.Context, ttype string, payload Payload) (err error) {
	log.Printf("Add New Task: {Type: %s, Payload: %#v}\n", ttype, payload)

	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = taskqueue.Add(*ctx, &taskqueue.Task{Payload: encoded, Method: "PULL", Tag: ttype}, PullQueueName)
	if err != nil {
		return err
	}

	// Any time we add new data task in a queue we need a worker task for it
	err = createTaskWorker(ctx)
	if err != nil {
		return err
	}

	return nil
}

func createTaskWorker(ctx *appengine.Context) (err error) {
	t := taskqueue.NewPOSTTask("/tasks/worker", url.Values{})
	_, err = taskqueue.Add(*ctx, t, PushQueueName)
	if err != nil {
		return err
	}

	return nil
}

func CreateWorker(ctx *appengine.Context) error {
	// Here we need to select tasks from PULL Queue based on it's operation type
	// Priority right now is following:
	//     1. Task on getting information on newly added user
	//     2. Task on regular user information pull
	//     3. Task on user information pull for pages deeper than first
	//
	// All this workers meant to be fired each 10 seconds
	// so we don't DDoS Instagram servers

	tasks, err := taskqueue.LeaseByTag(*ctx, 1, PullQueueName, 3600, FetchNewUser)
	if err != nil {
		return err
	}

	if tasks != nil && len(tasks) != 0 {
		processNewUserTasks(ctx, tasks)
	}

	tasks, err = taskqueue.LeaseByTag(*ctx, 1, PullQueueName, 3600, FetchOldUser)
	if err != nil {
		return err
	}

	if tasks != nil && len(tasks) != 0 {
		processOldUserTask(ctx, tasks)
	}

	return nil
}

func processNewUserTasks(ctx *appengine.Context, tasks []*taskqueue.Task) {
	for _, task := range tasks {
		p := Payload{}
		err := json.Unmarshal(task.Payload, &p)
		if err != nil {
			log.Printf("Unmarshal task payload problem: %s\n", err.Error())
		} else {
			// Actual request to Instagram
			data, err := GetDataFromInstagram(ctx, p.Username)
			if err != nil {
				log.Printf("Problem fetching [%s] instagram: %s", p.Username, err.Error())
			} else {
				// TODO: implement deep Instagram mining
				//if data.MoreAvailable {
				//	p.Page = 1
				//	err = AddTaskInQueue(ctx, FetchUserPage, p)
				//	if err != nil {
				//		log.Printf("ERROR: Create New Page Task problem: %s", err.Error())
				//		continue
				//	}
				//}

				err = SaveNewUser(ctx, p.Username, data)
				if err != nil {
					log.Printf("Problem adding user [%s] to the Database: %s", p.Username, err.Error())
				}
			}
		}

		log.Println("HERE")

		// We need to delete task as soon as we processed it
		if err = taskqueue.Delete(*ctx, task, PullQueueName); err != nil {
			log.Printf("Problem deleting task: %s", err.Error())
		}
	}
}

func processOldUserTask(ctx *appengine.Context, tasks []*taskqueue.Task) {
	for _, task := range tasks {
		p := Payload{}
		err := json.Unmarshal(task.Payload, &p)
		if err != nil {
			log.Printf("Unmarshal task payload problem: %s", err.Error())
		} else {
			// Actual request to Instagram
			data, err := GetDataFromInstagram(ctx, p.Username)
			if err != nil {
				log.Printf("Problem fetching [%s] instagram: %s", p.Username, err.Error())
			} else {
				index, err := SaveUserMissingPhotos(ctx, p.Username, data)

				if index+1 == len(data.Items) && data.MoreAvailable {
					// It means that we reach the end of the first page
					// Generally it means user added more than 20 photos
					// So we need to crawl second page

					p.Page++
					err = AddTaskInQueue(ctx, FetchUserPage, p)
					if err != nil {
						log.Printf("Create New Page Task problem: %s\n", err.Error())
					}
				}
			}
		}

		log.Println("HERE")

		// We need to delete task as soon as we processed it
		if err = taskqueue.Delete(*ctx, task, PullQueueName); err != nil {
			log.Printf("Problem deleting task: %s\n", err.Error())
		}
	}
}
