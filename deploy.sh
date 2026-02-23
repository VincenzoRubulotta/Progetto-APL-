#!/bin/bash

cd Server_Go/ 
go get google.golang.org/grpc\
go get google.golang.org/protobuf

cd ..
cd ScoponeClient/
dotnet add package Grpc.Net.Client
dotnet add package Google.Protobuf
dotnet add package Grpc.Tools

cd .. 

docker-compose up 



