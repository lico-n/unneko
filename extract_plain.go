package main

import (
	"bytes"
)

func extractPlainFile(neko *NekoData) *extractedFile {
	uncompressed := uncompressNeko(neko, newUntilFirstErrorCompleteCond())

	return &extractedFile{
		data:     uncompressed,
		fileExtension: ".lua",
	}
}

type untilFirstErrorCompleteCond struct{
	hasErrored bool
}

func newUntilFirstErrorCompleteCond() *untilFirstErrorCompleteCond {
	return &untilFirstErrorCompleteCond{}
}

func (c *untilFirstErrorCompleteCond) Complete(neko *NekoData, _ []byte) bool {
	return neko.FullyRead() || c.hasErrored
}

var knownFileEndings = [][]byte{
	[]byte("\n        return effects\n    "),
	[]byte("   sub_print_r(t,\"  \")\r\n    end\r\n    warn()\r\nend"),
	[]byte("LogInfo(\"box2d\", result)\r\n-- end\r\n\r\n"),
	[]byte("\"protocol.party.slightpartyredpoint\",\n}\n\n"),
}
var knownNonFileEndings = map[string]bool{
	"return C": true,
	"return {": true,
	"return function(": true,
	"return function(sceneobject": true,
	"return SceneDestro": true,
	"return TrapOpenTri": true,
	"return B": true,
	"return D": true,
}




func (c *untilFirstErrorCompleteCond) InterruptBlock(neko *NekoData, uncompressed []byte) bool {
	//if anotherFile == 6097 {
	//	fmt.Println("==================================================================")
	//	fmt.Println("==================================================================")
	//	fmt.Println("==================================================================")
	//	fmt.Println(string(uncompressed))
	//	fmt.Println("==================================================================")
	//	fmt.Println("==================================================================")
	//	fmt.Println("==================================================================")
	//}
	for _, v := range knownFileEndings {
		if len(uncompressed) > len(v) && bytes.Contains(uncompressed[len(uncompressed)-len(v):], v) {
			c.hasErrored = true
			return true
		}
	}


	if c.isTopLevelReturn(uncompressed) {
		c.hasErrored = true
		return true
	}



	return neko.FullyRead() || c.hasErrored
}

func (c *untilFirstErrorCompleteCond) isTopLevelReturn(uncompressed []byte) bool {
	lastReturn := bytes.LastIndex(uncompressed, []byte("\nreturn"))
	if lastReturn < 0 {
		return false
	}



	ending := bytes.TrimRight(uncompressed[lastReturn+1:], " \t")
	if ending[len(ending)-1] != '\n' {
		if knownNonFileEndings[string(ending)] {
			return false
		}

		newLineIndex := bytes.Index(ending, []byte("\n"))

		if newLineIndex > 0 && len(ending)-1  != newLineIndex{
			return false
		}


		if len(uncompressed[lastReturn+1:]) >= 8 {
			//fmt.Printf("suspected end: %s\n", uncompressed[lastReturn+1:])
			return true
		}
		return false
	}

	for i := lastReturn+1 ; i <len(uncompressed); i++ {
		c := uncompressed[i]

		if c != ' ' &&
			c != '\r' &&
			c != '\n' &&
			c != '_' &&
			!('0' <= c && c <= '9') &&
			!('a' <= c && c <= 'z') &&
			!('A' <= c && c <= 'Z') {
			return false
		}
	}

	return true
}

func (c *untilFirstErrorCompleteCond) UntilError() bool {
	return true
}

func (c *untilFirstErrorCompleteCond) RecordError()  {
	c.hasErrored = true
}

func (c *untilFirstErrorCompleteCond) IsValidUncompress(uncompressed []byte) bool {
	for _, v := range uncompressed {
		if v < 32  && v != '\r' && v != '\n' && v != '\t' && v != 8{
			return false
		}
	}
	return true
}