package main

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type Transfer struct {
	fileName string
	channel  chan []byte
}

func NewTransferManager() *TransferManager {
	return &TransferManager{
		transferMap: make(map[string]*Transfer),
	}
}

type TransferManager struct {
	sync.Mutex
	transferMap map[string]*Transfer
}

func (m *TransferManager) NewTransfer(fileName string) (string, *Transfer) {
	buf := make([]byte, 4)
	_, _ = rand.Read(buf)
	id := hex.EncodeToString(buf)

	m.transferMap[id] = &Transfer{
		fileName: fileName,
		channel:  make(chan []byte),
	}

	return id, m.transferMap[id]
}

func (m *TransferManager) GetTransfer(id string) *Transfer {
	m.Lock()
	defer m.Unlock()

	c, ok := m.transferMap[id]
	if !ok {
		return nil
	}

	delete(m.transferMap, id)

	return c
}
