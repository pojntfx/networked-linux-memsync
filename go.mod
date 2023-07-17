module github.com/pojntfx/networked-linux-memsync

go 1.20

require github.com/loopholelabs/userfaultfd-go v0.1.0

require golang.org/x/sys v0.4.0 // indirect

replace github.com/loopholelabs/userfaultfd-go => ../userfaultfd-go
