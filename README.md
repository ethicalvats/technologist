# Technologist Stack

## Requirements
- Docker
- Docker compose

## Front End
- VueJS
- Typescript
- Babel
- Jest
- Cypress
- Yarn

## REST API
- NodeJS
- TypeScript
- InversifyJS
- Grunt
- ExpressJS
- RAML

## Distributed Storage Engine
- Golang
- RAFT Consensus protocol

## Usage

` ./start.sh `

## How it works

RAFT library (lib-raft) launches 3 nodes listening on PORTS

` raft-n1: 9001, 9000 `

` raft-n2: 9002, 9000 `

` raft-n1: 9003, 9000 `

All the above raft nodes can communicate with each other via RPC on their respective ports.

Each raft node is also listening on PORT `9000` for incoming messages. They have a leader selected on start and can function with atleast 2 nodes running and supoorts fault tolerance of 1 node failure.

REST API runs on PORT `3000` and has the internal service send the items to RAFT engine. Raft engine has a Gateway to handle the incoming HTTP requests and forward it to all the nodes. And only leader node can accept the request. Which then stores the items in raw bytes to its in memory Log and propagates it to its followers returning the entire log back to the Gateway. Gateway responds back with the same log in JSON format back to the REST API which parses it to convert raw bytes into the domain model (Todo).

Frontend Application sends the captured user input to the REST API and gets back all the todos as the reponse of the same request. All todos are then listed in the UI.