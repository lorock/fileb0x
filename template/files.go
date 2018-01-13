package template

var filesTemplate = `{{buildTags .Tags}}// Code generated by fileb0x at "{{.Now}}" from config file "{{.ConfigFile}}" DO NOT EDIT.

package {{.Pkg}}
{{$Compression := .Compression}}

import (
  "bytes"
  {{if not .Spread}}{{if $Compression.Compress}}{{if not $Compression.Keep}}"compress/gzip"{{end}}{{end}}{{end}}
  "io"
  "net/http"
  "os"
  "path"
  
  "golang.org/x/net/webdav"
  "golang.org/x/net/context"
  
{{if .Updater.Enabled}}
  "crypto/sha256"
	"encoding/hex"
  "log"
  "path/filepath"
  "strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
{{end}}
)

var ( 
  // CTX is a context for webdav vfs
  {{exported "CTX"}} = context.Background()

  {{if .Debug}}
  {{exported "FS"}} = webdav.Dir(".")
  {{else}}
  // FS is a virtual memory file system
  {{exported "FS"}} = webdav.NewMemFS()
  {{end}}

  // Handler is used to server files through a http handler
  {{exportedTitle "Handler"}} *webdav.Handler

  // HTTP is the http file system
  {{exportedTitle "HTTP"}} http.FileSystem = new({{exported "HTTPFS"}})
)

// HTTPFS implements http.FileSystem
type {{exported "HTTPFS"}} struct {}

{{if (and (not .Spread) (not .Debug))}}
{{range .Files}}
// {{exportedTitle "File"}}{{buildSafeVarName .Path}} is "{{.Path}}"
var {{exportedTitle "File"}}{{buildSafeVarName .Path}} = {{.Data}}
{{end}}
{{end}}

func init() {
  if {{exported "CTX"}}.Err() != nil {
		panic({{exported "CTX"}}.Err())
	}

{{if not .Debug}}
{{if not .Updater.Empty}}
var err error
{{end}}

{{range $index, $dir := .DirList}}
  {{if and (ne $dir "./") (ne $dir "/") (ne $dir ".") (ne $dir "")}}
  err = {{exported "FS"}}.Mkdir({{exported "CTX"}}, "{{$dir}}", 0777)
  if err != nil && err != os.ErrExist {
    panic(err)
  }
  {{end}}
{{end}}
{{end}}

{{if (and (not .Spread) (not .Debug))}}
  {{if not .Updater.Empty}}
  var f webdav.File
  {{end}}

  {{if $Compression.Compress}}
  {{if not $Compression.Keep}}
  var rb *bytes.Reader
  var r *gzip.Reader
  {{end}}
  {{end}}

  {{range .Files}}
  {{if $Compression.Compress}}
  {{if not $Compression.Keep}}
  rb = bytes.NewReader({{exportedTitle "File"}}{{buildSafeVarName .Path}})
  r, err = gzip.NewReader(rb)
  if err != nil {
    panic(err)
  }

  err = r.Close()
  if err != nil {
    panic(err)
  }
  {{end}}
  {{end}}

  f, err = {{exported "FS"}}.OpenFile({{exported "CTX"}}, "{{.Path}}", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
  if err != nil {
    panic(err)
  }

  {{if $Compression.Compress}}
  {{if not $Compression.Keep}}
  _, err = io.Copy(f, r)
  if err != nil {
    panic(err)
  }
  {{end}}
  {{else}}
  _, err = f.Write({{exportedTitle "File"}}{{buildSafeVarName .Path}})
  if err != nil {
    panic(err)
  }
  {{end}}

  err = f.Close()
  if err != nil {
    panic(err)
  }
  {{end}}
{{end}}

  {{exportedTitle "Handler"}} = &webdav.Handler{
    FileSystem: FS,
    LockSystem: webdav.NewMemLS(),
  }

{{if .Updater.Enabled}}
  go func() {
    svr := &{{exportedTitle "Server"}}{}
    svr.Init()
  }()
{{end}}
}

// Open a file
func (hfs *{{exported "HTTPFS"}}) Open(path string) (http.File, error) {
  f, err := {{if .Debug}}os{{else}}{{exported "FS"}}{{end}}.OpenFile({{if not .Debug}}{{exported "CTX"}}, {{end}}path, os.O_RDONLY, 0644)
  if err != nil {
    return nil, err
  }

  return f, nil
}

// ReadFile is adapTed from ioutil
func {{exportedTitle "ReadFile"}}(path string) ([]byte, error) {
  f, err := {{if .Debug}}os{{else}}{{exported "FS"}}{{end}}.OpenFile({{if not .Debug}}{{exported "CTX"}}, {{end}}path, os.O_RDONLY, 0644)
  if err != nil {
    return nil, err
  }

  buf := bytes.NewBuffer(make([]byte, 0, bytes.MinRead))

  // If the buffer overflows, we will get bytes.ErrTooLarge.
  // Return that as an error. Any other panic remains.
  defer func() {
    e := recover()
    if e == nil {
      return
    }
    if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
      err = panicErr
    } else {
      panic(e)
    }
  }()
  _, err = buf.ReadFrom(f)
  return buf.Bytes(), err
}

// WriteFile is adapTed from ioutil
func {{exportedTitle "WriteFile"}}(filename string, data []byte, perm os.FileMode) error {
  f, err := {{exported "FS"}}.OpenFile({{exported "CTX"}}, filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
  if err != nil {
    return err
  }
  n, err := f.Write(data)
  if err == nil && n < len(data) {
    err = io.ErrShortWrite
  }
  if err1 := f.Close(); err == nil {
    err = err1
  }
  return err
}

// WalkDirs looks for files in the given dir and returns a list of files in it
// usage for all files in the b0x: WalkDirs("", false)
func {{exportedTitle "WalkDirs"}}(name string, includeDirsInList bool, files ...string) ([]string, error) {
	f, err := {{exported "FS"}}.OpenFile({{exported "CTX"}}, name, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	fileInfos, err := f.Readdir(0)
	if err != nil {
    return nil, err
  }
  
  err = f.Close()
  if err != nil {
		return nil, err
	}

	for _, info := range fileInfos {
		filename := path.Join(name, info.Name())

		if includeDirsInList || !info.IsDir() {
			files = append(files, filename)
		}

		if info.IsDir() {
			files, err = {{exportedTitle "WalkDirs"}}(filename, includeDirsInList, files...)
			if err != nil {
				return nil, err
			}
		}
	}

	return files, nil
}

{{if .Updater.Enabled}}
// Auth holds information for a http basic auth
type {{exportedTitle "Auth"}} struct {
  Username string
  Password string
}

// ResponseInit holds a list of hashes from the server
// to be sent to the client so it can check if there
// is a new file or a changed file
type {{exportedTitle "ResponseInit"}} struct {
  Success bool
  Hashes  map[string]string
}

// Server holds information about the http server
// used to update files remotely
type {{exportedTitle "Server"}} struct {
  Auth {{exportedTitle "Auth"}}
  Files []string
}

// Init sets the routes and basic http auth 
// before starting the http server
func (s *{{exportedTitle "Server"}}) Init() {
  s.Auth = {{exportedTitle "Auth"}}{
    Username: "{{.Updater.Username}}",
    Password: "{{.Updater.Password}}",
  }

  e := echo.New()
  e.Use(middleware.Recover())
  e.Use(s.BasicAuth())
  e.POST("/", s.Post)
  e.GET("/", s.Get)

  log.Println("fileb0x updater server is running at port 0.0.0.0:{{.Updater.Port}}")
  if err := e.Start(":{{.Updater.Port}}"); err != nil {
    panic(err)
  }
}

// Get gives a list of file names and hashes
func (s *{{exportedTitle "Server"}}) Get(c echo.Context) error {
  log.Println("[fileb0x.Server]: Hashing server files...")
  
  // file:hash
  hashes := map[string]string{}

  // get all files in the virtual memory file system
  var err error
  s.Files, err = {{exportedTitle "WalkDirs"}}("", false)
  if err != nil {
    return err
  }

  // get a hash for each file
  for _, filePath := range s.Files {
    f, err := FS.OpenFile(CTX, filePath, os.O_RDONLY, 0644)
    if err != nil {
      return err
    }

    hash := sha256.New()
    _, err = io.Copy(hash, f)
    if err != nil {
      return err
    }

    hashes[filePath] = hex.EncodeToString(hash.Sum(nil))
  }

  log.Println("[fileb0x.Server]: Done hashing files")
  return c.JSON(http.StatusOK, &ResponseInit{
    Success: true,
    Hashes: hashes,
  })
}

// Post is used to upload a file and replace 
// it in the virtual memory file system
func (s *{{exportedTitle "Server"}}) Post(c echo.Context) error {
  file, err := c.FormFile("file")
	if err != nil {
		return err
	}

  log.Println("[fileb0x.Server]:", file.Filename, "Found request to upload a file")

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()


  newDir := filepath.Dir(file.Filename)
  _, err = {{exported "FS"}}.Stat({{exported "CTX"}}, newDir)
  if err != nil && strings.HasSuffix(err.Error(), os.ErrNotExist.Error()) {
    log.Println("[fileb0x.Server]: Creating dir tree", newDir)
    list := strings.Split(newDir, "/")
    var tree string
    
    for _, dir := range list {
      if dir == "" || dir == "." || dir == "/" || dir == "./" {
        continue
      }

      tree += dir + "/"
      err = {{exported "FS"}}.Mkdir({{exported "CTX"}}, tree, 0777)
      if err != nil && err != os.ErrExist {
        log.Println("failed", err)
        return err
      }
    }
  }

  log.Println("[fileb0x.Server]:", file.Filename, "Opening file...")
  f, err := {{exported "FS"}}.OpenFile({{exported "CTX"}}, file.Filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
  if err != nil && !strings.HasSuffix(err.Error(), os.ErrNotExist.Error()) {
    return err
  }

  log.Println("[fileb0x.Server]:", file.Filename, "Writing file into Virutal Memory FileSystem...")
  if _, err = io.Copy(f, src); err != nil {
		return err
	}

  if err = f.Close(); err != nil {
    return err
  }

  log.Println("[fileb0x.Server]:", file.Filename, "Done writing file")
  return c.String(http.StatusOK, "ok")
}

// BasicAuth is a middleware to check if 
// the username and password are valid
// echo's middleware isn't used because of golint issues
func (s *{{exportedTitle "Server"}}) BasicAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			u, p, _ := c.Request().BasicAuth()
			if u != s.Auth.Username || p != s.Auth.Password {
				return echo.ErrUnauthorized
			}

			return next(c)
		}
	}
}
{{end}}
`
