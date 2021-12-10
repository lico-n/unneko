# unneko

An extractor for Revived Witch nekodata by Lico#6969


## Usage as binary 

Download from release page and run in terminal 

```
$ ./unneko -o ./output inputfile.nekodata
```

In case the inputfile is a patch nekodata the file extension must be `.patch.nekodata` 


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


