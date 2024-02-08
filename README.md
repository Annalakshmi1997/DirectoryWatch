**Go Program To Watch a Directory With Rest API And used framework Type is GIN **

Prerequisites to run the program 

Install Go Version go 1.20 

Change the current working directory to the location where you want the cloned directory (Clone the project in C:/ driver as we have given folder path to watch path := "C:/DirectoryWatch/TextFiles")

Command To Clone Repository 
	
 git clone https://github.com/Annalakshmi1997/DirectoryWatch.git

Open The Project in Visual Studio Code and run the below commands in terminal

1. go mod tidy  --> Since all dependencies are available in go.mod file. So just run this command.
2. go run main.go --> To run the project

Below Is the API Name with Parameter 

1. Start the Whatcher  (Use POST Method)
   
	http://localhost:8081/start-watcher
	{
		"MagicWord": "Word",
		"Status":true
	}
2. Stop The Watcher (Use POST Method)

   	http://localhost:8081/start-watcher
	{
		"Status":false
	}
    
3. Get History Details with Status like Write,Create,Delete,Rename and Magic Word count in given directory (use GET method)

 	http://localhost:8081/get-task-details




	
    

   	
