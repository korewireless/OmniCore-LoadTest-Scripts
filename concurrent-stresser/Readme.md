# Steps
1. Create many devices(For example 50000) with any registry and Suffix of device name as Stresser,ex:Stresser0,Stresser1 etc.Use a single key certificate file so its easier to manage.Generate a token for 10 hours and replace it in main.go line 98.Also replace the registry id and subscription id in main.go line 93.
2. Create Docker images from the dockerfiles
3. Run a single instance of Control Unit.
4. Run multiple pods of Stresser Unit as needed.Change config json and update the corresponding url for control unit , max clients per Stresser unit pod and the broker url.
5. Each stresser unit takes in a unique time frame from control unit and starts connected to the mqtt broker.