# ipfs-add [![GoDoc][1]][2] [![Build Status][3]][4] [![Go Report Card][5]][6] [![Coverage Status][7]][8]
           
[1]: https://godoc.org/github.com/dimchansky/ipfs-add?status.svg
[2]: https://godoc.org/github.com/dimchansky/ipfs-add
[3]: https://travis-ci.org/dimchansky/ipfs-add.svg?branch=master
[4]: https://travis-ci.org/dimchansky/ipfs-add
[5]: https://goreportcard.com/badge/github.com/dimchansky/ipfs-add
[6]: https://goreportcard.com/report/github.com/dimchansky/ipfs-add
[7]: https://codecov.io/gh/dimchansky/ipfs-add/branch/master/graph/badge.svg
[8]: https://codecov.io/gh/dimchansky/ipfs-add

Tool to add a file or directory to IPFS. It acts like `ipfs add` command, but you don't 
need IPFS node running on your local machine. By default it uses Infura nodes (`https://ipfs.infura.io:5001`)
to add file or directory to IPFS, all directories are added recursively.

## Usage

```
USAGE:
  ./ipfs-add [options] <path>...

ARGUMENTS

  <path>... - The path to a file to be added to ipfs.

OPTIONS

  -H	Include files that are hidden. Only takes effect on directory add.
  -node string
    	The url of IPFS node to use. (default "https://ipfs.infura.io:5001")
  -v	Print program version.
```

### Example

Adding current directory to IPFS:

```bash
> ipfs-add .
added QmctKt7CJDnmxdj7hRYXyqsLFMeEvpJt5qV6qdMprtcyop folder/1375 - Astronaut Vandalism - alt.txt
added QmXR6qCcJxy3P7TsqxodBgqMbSZCZBqdSNEmzHPzdfagub folder/1375 - Astronaut Vandalism - transcript.txt
added QmNTh4Er9bxYq6ULd4reHPkoPiPwVbXN8YqJHrnfkQy7RH folder/1375 - Astronaut Vandalism.png
added QmdGnC6rtZ7K7ERKnHuZCZcztbv9ZhBvsobHdgCowmX59F folder
```

You can now refer to the added directories or files in a gateway, like so:

- Folder

    https://ipfs.infura.io/ipfs/QmdGnC6rtZ7K7ERKnHuZCZcztbv9ZhBvsobHdgCowmX59F
    
- File in folder

    https://ipfs.infura.io/ipfs/QmdGnC6rtZ7K7ERKnHuZCZcztbv9ZhBvsobHdgCowmX59F/1375%20-%20Astronaut%20Vandalism.png

You can also use [any other IPFS gateway](https://ipfs.github.io/public-gateway-checker/) instead of 
`https://ipfs.infura.io/` like:

* https://ipfs.io/
* https://ipfs.eternum.io/