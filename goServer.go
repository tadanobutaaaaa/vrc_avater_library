package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kodeworks/golang-image-ico"
	"golang.org/x/image/draw"
	"gopkg.in/ini.v1"
)

type Root map[string][]map[string]Booth

type Booth struct {
    Id   string `json:"id"`
    Src  string `json:"src"`
}

func GoServer() {
    r := gin.Default()

    home, err := os.UserHomeDir()
    if err != nil {
        log.Fatal(err)
    }
    DownloadPath := filepath.Join(home, "Downloads")
    currentPath, _ := os.Getwd()
    currentAvatarsPath := filepath.Join(currentPath, "Avatars")
    currentImagesPath := filepath.Join(currentPath, "Images")

    fmt.Println("CurrentPath: ", currentPath)
    fmt.Println("DownloadPath: ", DownloadPath)

    r.Use(cors.New(cors.Config{
        AllowOrigins: []string{
            "https://accounts.booth.pm",
        },
        AllowMethods: []string{
            "POST",
            "OPTIONS",
        },
        AllowHeaders: []string{
            "Content-Type",
        },
        AllowOriginFunc: func(origin string) bool {
            return origin == "chrome-extension://hdfbpdpcecklifkgfdjegflfigfmjfib"
        },
    }))

    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": true,
        })
    })

    r.POST("/send/fileImages", func(c *gin.Context) {
        //保存する用のフォルダがない場合、フォルダを作成する
        if _, err := os.Stat("Avatars"); os.IsNotExist(err) {
            os.Mkdir("Avatars", 0750)
        }
        if _, err := os.Stat("Images"); os.IsNotExist(err) {
            os.Mkdir("Images", 0750)
        }

        var jsonData Root

        if err := c.ShouldBindJSON(&jsonData); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": err.Error(),
            })
            return
        }

        entries, err := os.ReadDir(DownloadPath)
        if err != nil {
            log.Fatal(err)
        }
        
        for _, entry := range entries {
            for key := range jsonData {
                for _, jsonEntry := range jsonData[key] {
                    for name, booth := range jsonEntry {
                        if entry.IsDir() && strings.Contains(name, entry.Name()) {
                            //サムネイル画像が保存されているフォルダがあるか確認する
                            inAvatarsFolder := filepath.Join(currentAvatarsPath, booth.Id)
                            _, err := os.Stat(inAvatarsFolder)
                            
                            //サムネイル画像が存在しない場合は、ダウンロードする
                            if err != nil {
                                os.Mkdir(inAvatarsFolder, 0750)

                                url := booth.Src
                                resp, err := http.Get(url)
                                if err != nil {
                                    fmt.Println("上手くダウンロードができませんでした")
                                    return
                                }
                                defer resp.Body.Close()

                                //サムネイル画像を保存する
                                jpgThumbnail := filepath.Join(currentImagesPath, booth.Id + ".jpg")

                                out, err := os.Create(jpgThumbnail)
                                if err != nil {
                                    fmt.Println("上手く保存ができませんでした")
                                    return
                                }
                                defer out.Close()
                                io.Copy(out, resp.Body)

                                icoThumbnail := filepath.Join(currentAvatarsPath, booth.Id)
                                icoThumbnail = filepath.Join(icoThumbnail, booth.Id + ".ico")

                                //icoファイルを作成する
                                file, err := os.Open(jpgThumbnail)
                                if err != nil {
                                    fmt.Println("上手く開けませんでした")
                                    return
                                }
                                defer file.Close()

                                img, _, err := image.Decode(file)
                                if err != nil {
                                    fmt.Println("上手くデコードできませんでした", err)
                                    return
                                }

                                resizedImg := image.NewRGBA(image.Rect(0, 0, 256, 256))
                                draw.NearestNeighbor.Scale(resizedImg, resizedImg.Bounds(), img, img.Bounds(), draw.Over, nil)

                                icoFile, err := os.Create(icoThumbnail)
                                if err != nil {
                                    fmt.Println("上手く保存ができませんでした")
                                    return
                                }

                                err = ico.Encode(icoFile, resizedImg)
                                if err != nil {
                                    fmt.Println("上手く保存ができませんでした")
                                    return
                                }

                                fmt.Printf("名前: %s, ID: %s, SRC: %s\n", name, booth.Id, booth.Src)
                                
                                //iniファイルに書き込む
                                desktopIniPath := filepath.Join(inAvatarsFolder, "desktop.ini")

                                cfg := ini.Empty()
                                cfg.Section(".ShellClassInfo").Key("IconResource").SetValue(fmt.Sprintf("\"%s.ico\",0", booth.Id))
                                err = cfg.SaveTo(desktopIniPath)
                                if err != nil {
                                    fmt.Println("上手く保存ができませんでした")
                                    return
                                }

                                exec.Command("attrib", "+h", desktopIniPath).Run()
                                exec.Command("attrib", "+s", inAvatarsFolder).Run()
                                exec.Command("attrib", "+h", icoThumbnail).Run()
                            }
                        os.Rename(filepath.Join(DownloadPath, entry.Name()), filepath.Join(inAvatarsFolder, entry.Name()))

                        //サーバーへの負荷対策
                        time.Sleep(1 * time.Second)
                        }
                    }
                }
            }
        }

        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": err.Error(),
            })
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "response": jsonData,
        })
        fmt.Println(jsonData)
    })

    r.Run(":8080")
}