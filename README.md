# i3-autoname
automatically names the i3 workspaces to the Windows inside them



# install + deps
go get github.com/mattn/go-sqlite3

go get github.com/ethragur/i3ipc-go

go get github.com/ethragur/i3-autoname


# run


go run github.com/ethragur/i3ipc-go/main.go


# insert icons

go run main.go -i -class=st-256color -icon=\uf120

go run main.go -i -class=firefox -icon=\uf269

