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

log-->server
log-->main
log-->rpc

com-->lsp
com-->rpc

lsp-->|dependency of|server
lsp-->rpc

rpc-->server

server-->main
``` 
