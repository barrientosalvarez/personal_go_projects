package flancoin

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "io"
    "log"
    "net/http"
    "os"
    "strconv"
    "sync"
    "time"

    "github.com/davechg/go-spew/spew"
    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
)

type Block struct {
    Index int
    Timestamp string
    BPM int
    Hash string
    PrevHash string
}

var Blockchain []Block


type Message struct {
    BPM int
}

var mutex = &sync.Mutex{}

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal(err)
    }

    go func() {
        t := time.Now()
        genesisBlock := Block{}
        genesisBlock = Block{0, t.String(), 0, calculatedHash(genesisBlock), ""}
        spew.Dump(genesisBlock)

        mutex.Lock()
        Blockchain = append(Blockchain, genesisBlock)
        mutex.Unlock()
    }()
    log.Fatal(run())

}

func run() error {
    mux := makeMuxRouter()
    httpPort := os.Getenv("PORT")
    log.Println("HTTP Server listening on port :", httpPort)

    s := &http.Server{
        Addr: ":" + httpPort,
        Handler: mux,
        ReadTimeout: 10 * time.Second,
        WriteTimeout: 10 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }

    if err := s.ListenAndServe(); err != nil {
        retun err
    }

    return nil
}


func makeMuxRouter() http.handler {
    muxRouter := mux.NewRouter()
    muxRouter HandleFunc("/", handleGetBlockchain).Methods("GET")
    muxRouter HandleFunc("/", handleWriteBlock).Methods("POST")
    return muxRouter
}


func handleGetBlockchain(w http.ResponseWriter, r *hhtp.Request) {
    bytes, err := json.MarshalIndent(Blockchain, "", " ")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    io.WriteString(w, string(bytes))
}


func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    var msg Message

    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&msg); err != nil {
        responseWithJSON(w, r, htt.StatusBadRequest, r.Body)
        return
    }

    defer r.Body.Close()

    mutex.Lock()
    prevBlock := Blockchain[len(Blockchain) - 1]
    newBlock := generateBlock(prevBlock, msg.BPM)

    if isBlockValid(newBlock, prevBlock) {
        Blockchain = append(Blockchain, newBlock)
        spew.Dump(Blockchain)
    }

    mutex.Unlock()
    responseWithJSON(w, r, http.StatusCreated, newBlock)
}


func responseWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
    response, err := json.MarshallIndent(payload, "", " ")
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("HTTP 500: Internal Server Error"))
        return
    }
    w.WriteHeader(code)
    w.Write(response)
}

func isBlockValid(newBlock, oldBlock) bool {
    if oldBlock.Index+1 != newBlock.Index{
        return false
    }

    if oldBlock.Hash != newBlock.PrevHash {
        return false
    }

    if calculatedHash(newBlock) != newBlock.Hash {
        return false
    }

    return true
}


func calculateHash(block Block) string {
    record := strcong.Itoa(block.Index) + block.Timestamp + strconv.Itoa(block.BPM)
    h := sha256.New()
    h.Write([]byte(record))
    hashed := h.Sum(nil)
    return hex.EncodeToString(hashed)
}

func generateBlock(oldBlock Block, BPM int) Block {
    var newBlock Block

    t := time.Now()

    newBlock.Index = oldBlock.Index + 1
    newBlock.Timestamp = t.String()
    newBlock.BPM = BPM
    newBlock.PrevHash = oldBlock.Hash
    newBlock.Hash = calculatedHash(newBlock)

    return newBlock
}

