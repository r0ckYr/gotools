package main

import (
    "fmt"
    "os"
    "bufio"
    "log"
    "net/http"
    "io"
    "sync"
    "regexp"
    "flag"
    "time"
)

func sendRequest(urls <-chan string, wg *sync.WaitGroup){
    defer wg.Done()
    client := http.Client{Timeout: 10 * time.Second} 
    for url:= range urls{
        response, err := client.Get(url)
        if err != nil{
            fmt.Printf("Cannot Download %s: %s\n", url, err)
            continue
        }
        statusCode := response.StatusCode
       
        if statusCode != 200{
            continue
        }

        filename := generateFileName(url)
        
        jsfile, err := os.Create(filename)
        if err != nil{
            fmt.Printf("Cannot save %s : %s\n", url, err)
            response.Body.Close()
            continue
        }
        
        bytes, err := io.Copy(jsfile, response.Body)
        jsfile.Close()
        if err != nil{
            fmt.Printf("Cannot save %s : %s", filename, err)
            response.Body.Close()
            continue
        }
        response.Body.Close()
        fmt.Printf("%s [%d] [%d]\n",url, statusCode, bytes)
    }
}


func generateFileName(url string) string{
    url = regexp.MustCompile(`[^a-zA-Z0-9 ]+`).ReplaceAllString(url, "_")
    extension := url[len(url)-3:]
    if len(url) > 100{
        url = url[:101]
    }
    if extension == "_js"{
        url = url[:len(url)-3]           
        url += ".js"
    }else{
        url += ".js"
    }

    return url
}

func main(){
    threads := flag.Int("n", 50, "# number of threads")
    flag.Parse()
    if len(os.Args) < 2{
        fmt.Println("Usage: ./jshunter <urls.txt>")
        fmt.Println("Usage: ./jshunter -n 10 <urls.txt>")
        return
    }

    file_path := string(os.Args[len(os.Args)-1])
    file, err := os.Open(file_path)
    
    if err != nil{
        log.Fatal(err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)

    var urls []string

    for scanner.Scan(){
        urls = append(urls, scanner.Text())
    }

    maxConcurrency := *threads
    urlChan := make(chan string,len(urls)*maxConcurrency)    
    var wg sync.WaitGroup
    
    for _, url := range urls{
        urlChan <- url   
    }
    close(urlChan)

    for i:=0;i<maxConcurrency;i++ {
        wg.Add(1)
        go sendRequest(urlChan, &wg)
    }

    wg.Wait()
    fmt.Println("All file downloaded")
}
