# Packages

```mermaid
---
title: Packages
---
flowchart TD

com
log

lsp
rpc
server
main
analysis

log-->server
log-->main
log-->rpc
log-->analysis

com-->lsp
com-->rpc

lsp-->|dependency of|server
lsp-->rpc

rpc-->server

server-->main

analysis-->server
``` 
