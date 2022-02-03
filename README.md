# Chrome-Password-Dumper
Chrome password dumper written in Go for Linux and Windows.

# Usage
Dump all credentials to "out.json":   
`cpd_x64.exe -o out.json`

Pretty print all credentials to the console in JSON format:   
`cpd_x64.exe -p`

```
Usage: cpd_x64.exe [--outpath OUTPATH] [--print]

Options:
  --outpath OUTPATH, -o OUTPATH
                         Where to write credentials to. File extension must be ".txt" or ".json". Path will be made if it doesn't already exist.
  --print, -p            Write JSON to stdout.
  --help, -h             display this help and exit
```

# Output
JSON:
```json
[
    {
		"url": "https://app.napster.com/",
		"username": "x@gmail.com",
		"password": "x"
	},
	{
		"url": "https://account.napster.com/",
		"username": "x@gmail.com",
		"password": "x"
	},
	{
		"url": "https://www.duolingo.com/",
		"username": "x",
		"password": "x"
	}
]
```
Plain text:
```
URL: https://app.napster.com/
Username: x@gmail.com
Password: x

URL: https://account.napster.com/
Username: x@gmail.com
Password: x

URL: https://www.duolingo.com/
Username: x
Password: x
```
