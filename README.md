# Distribute In Memory Cache

# Installation
To run this application, we need to have go1.18 or above.

# Running
Here are the steps to run the application
- Remove `node` binary from node-binary directory
- Build the node server inside the **node-binary** directory
  - ```go build -o node ./...```
- Modify the `config.yaml` inside the **config** directory
- Run the `main.go` in the root directory 
  - ```go run main.go```
- Access the webui at `http://localhost:3000/` considering the master is running on *3000*

# Testing
Assuming master running in `3000` and one of the node running in `3001`

## Data Test
- Once the log prints `server is ready`, check the node statistics
  - In the web ui: **Node statistics** tab or 
  - Via rest api
    - ```curl -XGET http://localhost:3000/api/infra/nodestats```
  - This api show the available nodes and what their status are
- Push some data via `Set` api like this
  - In the webui: **Set data** tab, or
  - Via rest api
    - ```curl -XPOST http://localhost:3000/api/data/set -d '{"key2":"value"}'```
  - This will save the data in master and master will broadcast to all th nodes
- Check the data
  - In the webui: **View data** tab or
  - Via rest api
      - In master server
        - ```curl -XGET http://localhost:3000/api/data/get```
      - In slaves
        - ```curl -XGET http://localhost:3001/data```
        - ```curl -XGET http://localhost:3002/data``` etc
- Do the same with data via `Delete` api or in webui, via the **Delete data** tab

## Node recovery
- Kill the master node
- Check the slaves still running by checking any endpoint like this
  - ```curl -XGET http://localhost:3001/health```
- Restart the master server, it will reconnect the nodes and sync itself with the updated slaves data