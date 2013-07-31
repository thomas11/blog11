#!/bin/sh

rm -rf out/*
cp -R static out/
go run blog11.go
