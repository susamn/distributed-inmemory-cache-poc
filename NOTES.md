# Introduction
This is a rudimentary in-memory distributed cache implementation in Golang.

# Architecture
This simple solution has 1 master server and multiple node servers. They communicate via http protocol and its a two 
way communication. When data is changed(set or delete), master server sends broadcast signal with a new data version id.
The nodes compare theirs data version id and if they are out of date, they sync themselves with the new data from master.
If the master goes down, the next time it comes up, it will recover any nodes already running and sync itself with the most
upto date from any node based on the data version.

The implementation is divided into 4 components. 
- Config
- The master server component
- The node component
- Web based data manager

Due to the limited scope and time, there are a number of things left unimplemented, such as:
- Communication from master and nodes must be made securely. **mTLS** can be used in both master and nodes to make them identify themselves
- The UI is very simple is not intuitive, but it serves the purpose.
- The auto scale up and scale down can be achieved by another daemon process periodically
- The master now acts as the data read and write endpoint. We can create separate read endpoint which will read from all the nodes and serve, rather
than hogging the master server, the writes can go via master.
- I have not witten any test cases to test the apis
- The nodes are running in the same physical machines. It is possible to run the nodes in another physical machine. Separate code, some kind of agent is needed.
- The accessing webui or killing nodes or scaling must be done under strict access management, which must be added. Only master should be able to start or stop nodes.


## Config
The config directory in the root acts as the configuration for the project. It contains a single `config.yaml` file,
which is read when starting the master server. We can modify the config to match our setup.

## Web based data manager
Once the server

## Master server
The master server is started on the port mentioned in the config. The master server will spawn the configured numbers
of nodes as daemon process. If the master goes down, still the nodes will be running. The nodes will have port number
starting from `node_port_initial` and counting up.

### The slave wrapper
The master server uses a separate struct `slave.go` to keep track of the nodes running. The `slave.go` is kind of a 
wrapper around the running daemon nodes. It's a way to control and tackle the detached processes(nodes) which the master
spawns but are detached from the master server to make the nodes truly independent. This will also make it possible to 
run nodes in another physical servers.

## Node server
The **node-binary** folder acts as the node server component. It has a precompiled binary called **node** (Compiled for debian x64) which 
is used by master server to spawn daemon node server components. For running in another machine, it must be recompiled for that machine using
`go build -o node ./...` inside the **node-binary** folder.

## Web based data manager
The web based data manager is available at `http://localhost:3000/` considering the master is running in *3000*. With the web app,
data set, get, delete can be performed. It is built with vue3 and the compiled web resources are inside the **public** directory which is 
served by the master server. The source code for the web is available in the **web** directory.
