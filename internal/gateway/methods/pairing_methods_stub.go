package methods

import (
	"context"

	"github.com/nextlevelbuilder/goclaw/internal/gateway"
	"github.com/nextlevelbuilder/goclaw/internal/store"
	"github.com/nextlevelbuilder/goclaw/pkg/protocol"
)

// PairingMethods stub - pairing removed
type PairingMethods struct {
	pairingStore store.PairingStore
}

// NewPairingMethods stub
func NewPairingMethods(store store.PairingStore) *PairingMethods {
	return &PairingMethods{pairingStore: store}
}

// SetBroadcaster stub
func (m *PairingMethods) SetBroadcaster(broadcaster func(event protocol.EventFrame)) {
	// no-op
}

// Register stub
func (m *PairingMethods) Register(router *gateway.MethodRouter) {
	// pairing methods removed
}

func (m *PairingMethods) handleListPairings(_ context.Context, client *gateway.Client, req *protocol.RequestFrame) {
	client.SendResponse(protocol.NewOKResponse(req.ID, map[string]interface{}{"pairings": []interface{}{}}))
}
