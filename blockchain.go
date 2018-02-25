package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

//Block is struct which contains all data
type Block struct {
	Index     int
	Timestamp string
	Money     int
	Hash      string
	PrevHash  string
}

//Blockchain contains blocks
var Blockchain []Block

//generateHash generating hash for new blocks
func generateHash(block Block) string {
	create := string(block.Index) + block.Timestamp + string(block.Money) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(create))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

//generateBlock generating new blocks
func generateBlock(oldBlock Block, Money int) (Block, error) {

	var newBlock Block

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = time.Now().String()
	newBlock.Money = Money
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = generateHash(newBlock)

	return newBlock, nil
}

//isBlockValid checking if block is valid
func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index || oldBlock.Hash != newBlock.PrevHash || generateHash(newBlock) != newBlock.Hash {
		return false
	}
	return true
}

//replaceChain replace chain if it longer than blockchain
func replaceChain(newBlocks []Block) {
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
}

//handleGetBlockchain sending json data with blockchain information
func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(Blockchain, "", " ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}

//Message contains data from user
type Message struct {
	Money int
}

//respondWithJSON sending json data to browser
func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

//handleWriteBlock handle request on create new block
func handleWriteBlock(w http.ResponseWriter, r *http.Request) {

	var m Message

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()

	newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], m.Money)
	if err != nil {
		respondWithJSON(w, r, http.StatusInternalServerError, m)
		return
	}
	if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
		newBlockchain := append(Blockchain, newBlock)
		replaceChain(newBlockchain)
	}
	respondWithJSON(w, r, http.StatusCreated, newBlock)
}

//makeMuxRouter create routes
func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
	return muxRouter
}

//run start server
func run() error {
	mux := makeMuxRouter()
	httpAddr := "8080"
	log.Println("Listening on ", httpAddr)
	s := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	return s.ListenAndServe()
}

func main() {

	go func() {

		//create template for genesis block
		template := Block{0, time.Now().String(), 0, "", ""}

		//create genesis block
		genesisBlock := Block{0, time.Now().String(), 0, generateHash(template), ""}

		//append genesis block to blockchain
		Blockchain = append(Blockchain, genesisBlock)
	}()

	//start server
	run()
}
