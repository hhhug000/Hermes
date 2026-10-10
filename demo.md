# Hermes Demo

Hello! Do you want to try out Hermes?

Well then you're in the right place!

Just follow the steps below to try it out!

## Demo video

If you dont want to bother setting up a Hermes server, you can just watch this video to see how it works (Kinda boring but still cool)

[Demo Video](https://www.loom.com/share/2c7bc76073da47b3be5ddd8da2aae2e3)

## Try it out yourself

If you actually want to use Hermes yourself, then just run the commands below to install and run it:

```
git clone https://github.com/hhhug000/Hermes.git
cd Hermes
cp hermes.conf.example hermes.conf
go run .
```

Then to use it you can use these commands:

### Open the TUI

```
ssh localhost -p 2222
```

### Send a file to Hermes

```
echo file.txt | ssh localhost -p 2222 share file.txt
ssh localhost -p 2222
```

That second bit is to activate and share the file

### Download a file from Hermes

```
ssh localhost -p 2222 download fileId > file.txt
```

The fileId can be found in the TUI
fileIds can be a bit hard to remember, so the right command will be given when you select it in the TUI
