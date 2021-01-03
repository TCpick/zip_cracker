package main

import (
        "bytes"
        "fmt"
        "io"
        "log"
        "os"
        "strings"
        "time"

        //"github.com/yeka/zip"
        "github.com/alexmullins/zip"
        "runtime/pprof"
)

const (
        THREAD = 100
)

var c_password = make(chan string, THREAD)
var c_break = make(chan string)
var startTime time.Time
var count int = 0

func PasswordGen(alphabet []string) {
        for i := 1; i <= 10; i++ {
                for combo := range GenerateCombinationsString(alphabet, i) {
                        c_password <- strings.Join(combo, "")
                        count++
                        if count & 0xfffff == 0 {
                                fmt.Printf("1 Mega have been tried. Now %s, Time taken: %f seconds\n",
                                        strings.Join(combo, ""), time.Since(startTime).Seconds())
                        }
                }
        }
        c_break <- "password exhausted!!!"
}

func GenerateCombinationsString(data []string, length int) <-chan []string {
        c := make(chan []string)
        go func(c chan []string) {
                defer close(c)
                combosString(c, []string{}, data, length)
        }(c)
        return c
}

func combosString(c chan []string, combo []string, data []string, length int) {
        if length <= 0 {
                return
        }
        var newCombo []string
        for _, ch := range data {
                newCombo = append(combo, ch)
                if length == 1 {
                        output := make([]string, len(newCombo))
                        copy(output, newCombo)
                        c <- output
                }
                combosString(c, newCombo, data, length-1)
        }
}

func unzip(filename string, password string) {
        r, err := zip.OpenReader(filename)
        if err != nil {
                return
        }
        defer r.Close()

        for _, f := range r.File {
                f.SetPassword(password)
                r, err := f.Open()
                if err != nil {
                        return
                }
                defer r.Close()
                buffer := new(bytes.Buffer)
                n, err := io.Copy(buffer, r)
                if n == 0 || err != nil {
                        return
                }
                break
        }
        c_break <- fmt.Sprintf("password found!!! %s", password)
        return
}

func bruteforce(zipFile string, alphabet []string) {
        startTime = time.Now()
        go PasswordGen(alphabet)
        c_timeout := time.After(time.Second*20)

LOOP:
        for {
                select {
                case password := <-c_password:
                        go unzip(zipFile, password)
                case info := <-c_break:
                        fmt.Printf("Process break due to: %s\n", info)
                        break LOOP
                case <-c_timeout:
                        fmt.Printf("Process break due to timeout\n")
                        break LOOP
                }
        }

        fmt.Printf("Program Finished: Combinations tried: %d, Time taken: %f seconds\n",
                count, time.Since(startTime).Seconds())
}

func main() {
        // profiling part
        fprof, err := os.Create("cpuProfiling")
        if err != nil {
                log.Fatal(err)
        }
        pprof.StartCPUProfile(fprof)
        defer pprof.StopCPUProfile()
        // ***profiling part***

        if len(os.Args) < 4 {
                fmt.Printf("\nUsage: %s [zip file] [letters] [type of attack]\n\nExample:\n\t- Brute force: %s ExampleFile.zip abcdefghijklmnopqrstuvwxyz bruteforce\n\n", os.Args[0], os.Args[0])
                os.Exit(1)
        }

        zipFile := os.Args[1]
        chars := os.Args[2]
        attack := os.Args[3]

        if attack == "bruteforce" {
                fmt.Println("Starting brute force attack..")
                alphabet := strings.Split(chars, "")
                bruteforce(zipFile, alphabet)
        } else {
                os.Exit(1)
        }
}
