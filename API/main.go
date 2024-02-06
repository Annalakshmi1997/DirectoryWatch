package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/radovskyb/watcher"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Declare structure
type DirectoryDetails struct {
	FileName        string    `json:"FileName" bson:"FileName"`
	TaskStartTime   time.Time `json:"TaskStartTime" bson:"TaskStartTime"`
	TaskEndTime     time.Time `json:"TaskEndTime" bson:"TaskEndTime"`
	TaskDuration    string    `json:"TaskDuration" bson:"TaskDuration"`
	TaskName        string    `json:"TaskName" bson:"TaskName"`
	TaskStatus      string    `json:"TaskStatus" bson:"TaskStatus"`
	MagicWord       string    `json:"MagicWord" bson:"MagicWord"`
	Count           int       `json:"Count" bson:"Count"`
	CreatedDateTime time.Time `json:"CreatedDateTime" bson:"CreatedDateTime"`
}

var (
	status      bool
	statusMutex sync.Mutex
)

// any modifications made to the fields within the method will directly affect the original instance
func (d *DirectoryDetails) AssignDataToFields(fileName string, startTime, endTime time.Time, TimeDuration string, taskName string, taskstatus string, count int, word string, createdDateTime time.Time) {
	d.FileName = fileName
	d.TaskStartTime = startTime
	d.TaskEndTime = endTime
	d.TaskDuration = TimeDuration
	d.TaskName = taskName
	d.TaskStatus = taskstatus
	d.MagicWord = word
	d.Count = count
	d.CreatedDateTime = createdDateTime
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	//Handle CORS policy
	config := cors.DefaultConfig()
	config.AllowMethods = []string{"*"}
	config.AllowOrigins = []string{"http://localhost:8080"}
	router.Use(cors.New(config))
	router.POST("/start-watcher", watcherstart)
	router.POST("/stop-watcher", stopwatcher)
	router.GET("/get-task-details", GetTaskDetailsHandler)
	fmt.Println("Watching for changes...")
	//Set default port number
	router.Run(":8081")

}

// API To Start Watcher
func watcherstart(c *gin.Context) {
	//Struct to bind(decode) json value
	type MagicString struct {
		MagicWord string
		Status    bool
	}
	newConfig := MagicString{}
	if err := c.BindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}
	//Initiate New Watcher
	w := watcher.New()
	w.FilterOps(watcher.Write, watcher.Create, watcher.Remove, watcher.Rename)
	//Declare and add the Directory Path to watch
	path := "C:/DirectoryWatch/TextFiles"
	if err := w.Add(path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Msg": err})
		return
	}
	//Implement Go routine function to get values from channels. Will run when write/delete/rename/create event occurs in specified folder
	go func() {
		for {
			select {
			case event := <-w.Event:
				statusMutex.Lock()
				//Close the watcher based on status value
				if status {
					w.Close()
					status = false
				} else {
					//Call save and CountMagigString function to save event with magic string count in backend
					startTime := time.Now()
					fmt.Println("Check Event", event)
					count := CountMagicString(newConfig.MagicWord, path)
					Status, ErrorString := Save(event, startTime, count, newConfig.MagicWord)
					if Status {
						c.JSON(http.StatusInternalServerError, gin.H{"Msg": ErrorString})
						return
					}

				}
				statusMutex.Unlock()
			case err := <-w.Error:
				log.Fatal(err)
			case <-w.Closed:
				//If watcher is closed, then
				return
			}
		}
	}()
	//c.JSON(http.StatusOK, gin.H{"Msg": "Watcher S Successfully"})
	//Start the watcher every 100 Milliseconds
	if err := w.Start(time.Millisecond * 100); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Msg": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"Msg": "Watcher Stoped Successfully"})
}
func stopwatcher(c *gin.Context) {
	status = true
	c.JSON(http.StatusOK, gin.H{"Msg": "Watcher Stoped Successfully"})
}

// Get Task List and Magic Word Count Occurrence Count
func GetTaskDetailsHandler(c *gin.Context) {
	clientoption := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ := mongo.Connect(context.Background(), clientoption)
	collection := client.Database("DirectoryWatcher").Collection("TaskDetails")
	filter := bson.M{}
	output := []bson.M{}
	cur, err2 := collection.Find(context.Background(), filter)
	if err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data not found"})
		return
	}
	e := cur.All(context.Background(), &output)
	defer cur.Close(context.Background())
	if e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data not found"})
		return
	}
	c.JSON(http.StatusOK, output)
}
func CountMagicString(MagicText string, path string) int {
	// Declare Wait Group to implement Wait Group Methods
	var wg sync.WaitGroup
	count := 0
	//Read the dirctory of given path
	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		wg.Add(1)

		if file.IsDir() {
			// Skip directories if needed
			continue
		}

		// Construct the full file path
		filePath := filepath.Join(path, file.Name())
		// implement a go routine function to get count
		go fileprocess(filePath, &count, &wg, MagicText)

	}
	wg.Wait()
	return count

	//return occurrences, nil
}
func fileprocess(filePath string, count *int, wg *sync.WaitGroup, MagicText string) {
	//Make wait group internal count zero
	defer wg.Done()
	// Read the file content of file
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Println("Error reading file:", err)
		return
	}
	// Count the magic string occurrence in file
	*count += strings.Count(string(content), MagicText)
}
func Save(event watcher.Event, StartTime time.Time, count int, word string) (bool, string) {
	fmt.Println("Check Event Values", event)
	//Binding Db URL
	clientoption := options.Client().ApplyURI("mongodb://localhost:27017")
	//Connect with Mongodb By using Above URL
	client, err := mongo.Connect(context.Background(), clientoption)
	if err != nil {
		return true, "Error In Db Connection"
	}
	collection := client.Database("DirectoryWatcher").Collection("TaskDetails")
	SaveTask := DirectoryDetails{}
	endTime := time.Now()
	duration := endTime.Sub(StartTime)
	hours := int(duration.Hours())
	Min := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	TimeDuration := fmt.Sprintf("%d hours %d Minutes %d Sec", hours, Min, seconds)
	//Calling method to assign field values
	SaveTask.AssignDataToFields(event.FileInfo.Name(), time.Now(), endTime, TimeDuration, event.Op.String(), "Completed", count, word, time.Now())
	//Save Record in mongo db
	_, err1 := collection.InsertOne(context.Background(), SaveTask)
	if err1 != nil {
		return true, "Error in Data Insert in Backend"
	}
	return false, ""
}
