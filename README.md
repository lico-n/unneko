# unneko

An extractor for Revived Witch nekodata by Lico#6969


## Usage as command line binary 

Download unneko binary from release page and run in terminal 

```
single nekodata file
$ ./unneko -o ./output inputfile.nekodata

or directory containing nekodata files
$ ./unneko -o ./output ./input/
```

In case the inputfile is a patch nekodata the file extension must be `.patch.nekodata` 

## Usage with windows GUI

For those on Windows and unfamiliar with basic command line usage, you can download `unneko-win-x64.exe`.
Then simply drag and drop nekodata file or a directory containing nekodata files on the `unneko-win-x64.exe`, 
the extracted files will appear in `output` directory.

It is possible that windows/smart defender will pop up a warning. This is a common false positive because 
the windows binary is unsigned. In this case press more information and run anyways.

## Usage as library

```go
import "github.com/lico-n/unneko"

func main() {
  // initialize with nekodata filepath, if it's patch nekodata must contain `.patch.nekodata` file extension
  r, err := unneko.NewReaderFromFile(inputFilePath)
  handleError(err)

  // alternatively initialize with in memory nekodata, must provide flag whether this is a patch file
  var inMemoryNekoData []byte
  isPatchFile := true 

  r, err = unneko.NewReader(inMemoryNekoData, isPatchFile)
  handleError(err)

  // iterate over all extracted files
  for r.HasNext() {
    file, err := r.Next()
    handleError(err)

    file.Path() // original file path of the extracted file 
    file.Data() // extracted data as byte array 
  }

}
```


