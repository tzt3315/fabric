/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deliver_test

import (
	"io"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric/common/deliver"
	"github.com/hyperledger/fabric/common/deliver/mock"
	"github.com/hyperledger/fabric/common/ledger/blockledger"
	"github.com/hyperledger/fabric/common/util"
	cb "github.com/hyperledger/fabric/protos/common"
	ab "github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/pkg/errors"
	"golang.org/x/net/context"

<<<<<<< HEAD
var sendDeliverResponseProducer = func(srv ab.AtomicBroadcast_DeliverServer) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		return srv.Send(msg.(*ab.DeliverResponse))
	}
}

var timeWindow = time.Duration(15 * time.Minute)
var testCert = &x509.Certificate{
	Raw: []byte("test"),
}

const ledgerSize = 10
const mutualTLS = true

func init() {
	flogging.SetModuleLevel(pkgLogID, "DEBUG")
	factory.InitFactories(nil)
}

type mockStream struct {
	grpc.ServerStream
	StreamContext context.Context
	CancelContext func()
}

func (m mockStream) Context() context.Context {
	if m.StreamContext != nil {
		return m.StreamContext
	}
	p := &peer.Peer{}
	p.AuthInfo = credentials.TLSInfo{
		State: tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{
				testCert,
			},
		},
	}
	return peer.NewContext(context.Background(), p)
}

type mockD struct {
	mockStream
	recvChan chan *cb.Envelope
	sendChan chan *ab.DeliverResponse
}

func (m *mockD) CreateStatusReply(status cb.Status) proto.Message {
	return &ab.DeliverResponse{
		Type: &ab.DeliverResponse_Status{Status: status},
	}
}
=======
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)
>>>>>>> 95483dd862a768c588c2582e88f6b9da37a47ed7

var (
	seekOldest = &ab.SeekPosition{
		Type: &ab.SeekPosition_Oldest{Oldest: &ab.SeekOldest{}},
	}

<<<<<<< HEAD
func newMockD() *mockD {
	p := &peer.Peer{}
	p.AuthInfo = credentials.TLSInfo{
		State: tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{
				testCert,
			},
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	peer.NewContext(ctx, p)

	return &mockD{
		mockStream: mockStream{
			StreamContext: ctx,
			CancelContext: cancel,
		},
		recvChan: make(chan *cb.Envelope),
		sendChan: make(chan *ab.DeliverResponse),
=======
	seekNewest = &ab.SeekPosition{
		Type: &ab.SeekPosition_Newest{Newest: &ab.SeekNewest{}},
>>>>>>> 95483dd862a768c588c2582e88f6b9da37a47ed7
	}
)

var _ = Describe("Deliver", func() {
	Describe("NewHandler", func() {
		var fakeChainManager *mock.ChainManager

		BeforeEach(func() {
			fakeChainManager = &mock.ChainManager{}
		})

		It("returns a new handler", func() {
			handler := deliver.NewHandler(fakeChainManager, time.Second, false)
			Expect(handler).NotTo(BeNil())

			Expect(handler.ChainManager).To(Equal(fakeChainManager))
			Expect(handler.TimeWindow).To(Equal(time.Second))
			// binding inspector is func; unable to easily validate
			Expect(handler.BindingInspector).NotTo(BeNil())
		})
	})

	Describe("ExtractChannelHeaderCertHash", func() {
		It("extracts the TLS certificate hash from a channel header", func() {
			chdr := &cb.ChannelHeader{TlsCertHash: []byte("tls-cert")}

			result := deliver.ExtractChannelHeaderCertHash(chdr)
			Expect(result).To(Equal([]byte("tls-cert")))
		})

		Context("when the message is not a channel header", func() {
			It("returns nil", func() {
				result := deliver.ExtractChannelHeaderCertHash(&cb.Envelope{})
				Expect(result).To(BeNil())
			})
		})

		Context("when the message is nil", func() {
			It("returns nil", func() {
				var ch *cb.ChannelHeader
				result := deliver.ExtractChannelHeaderCertHash(ch)
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("Handle", func() {
		var (
			errCh              chan struct{}
			fakeChain          *mock.Chain
			fakeBlockReader    *mock.BlockReader
			fakeBlockIterator  *mock.BlockIterator
			fakeChainManager   *mock.ChainManager
			fakePolicyChecker  *mock.PolicyChecker
			fakeReceiver       *mock.Receiver
			fakeResponseSender *mock.ResponseSender
			fakeInspector      *mock.Inspector

			handler *deliver.Handler
			server  *deliver.Server

			channelHeader *cb.ChannelHeader
			seekInfo      *ab.SeekInfo
			ts            *timestamp.Timestamp

			channelHeaderPayload []byte
			seekInfoPayload      []byte
			envelope             *cb.Envelope
		)

		BeforeEach(func() {
			errCh = make(chan struct{})
			fakeChain = &mock.Chain{}
			fakeChain.ErroredReturns(errCh)

			block := &cb.Block{
				Header: &cb.BlockHeader{Number: 100},
			}
			fakeBlockIterator = &mock.BlockIterator{}
			fakeBlockIterator.NextReturns(block, cb.Status_SUCCESS)

			fakeBlockReader = &mock.BlockReader{}
			fakeBlockReader.HeightReturns(1000)
			fakeBlockReader.IteratorReturns(fakeBlockIterator, 100)
			fakeChain.ReaderReturns(fakeBlockReader)

			fakeChainManager = &mock.ChainManager{}
			fakeChainManager.GetChainReturns(fakeChain, true)

			fakePolicyChecker = &mock.PolicyChecker{}
			fakeReceiver = &mock.Receiver{}
			fakeResponseSender = &mock.ResponseSender{}

			fakeInspector = &mock.Inspector{}

			handler = &deliver.Handler{
				ChainManager:     fakeChainManager,
				TimeWindow:       time.Second,
				BindingInspector: fakeInspector,
			}
			server = &deliver.Server{
				Receiver:       fakeReceiver,
				PolicyChecker:  fakePolicyChecker,
				ResponseSender: fakeResponseSender,
			}

			ts = util.CreateUtcTimestamp()
			channelHeader = &cb.ChannelHeader{
				ChannelId: "chain-id",
				Timestamp: ts,
			}
			seekInfo = &ab.SeekInfo{
				Start: &ab.SeekPosition{
					Type: &ab.SeekPosition_Specified{
						Specified: &ab.SeekSpecified{Number: 100},
					},
				},
				Stop: &ab.SeekPosition{
					Type: &ab.SeekPosition_Specified{
						Specified: &ab.SeekSpecified{Number: 100},
					},
				},
			}

<<<<<<< HEAD
	select {
	case deliverReply := <-mockSrv.sendChan:
		if deliverReply.GetStatus() != cb.Status_SUCCESS {
			t.Fatalf("Expected delivery to complete")
		}
	case <-time.After(time.Second):
		t.Fatalf("Timed out waiting to get all blocks")
	}
}

func TestErroredSeek(t *testing.T) {
	mm := newMockMultichainManager()
	ms := mm.chains[systemChainID]
	l := ms.ledger
	close(ms.erroredChan)
	for i := 1; i < ledgerSize; i++ {
		l.Append(blockledger.CreateNextBlock(l, []*cb.Envelope{{Payload: []byte(fmt.Sprintf("%d", i))}}))
	}

	mockSrv := newMockD()
	defer close(mockSrv.recvChan)
	ds := initializeDeliverHandler(mm, !mutualTLS, false)
	m := NewDeliverServer(mockSrv, mm.PolicyChecker, sendDeliverResponseProducer(mockSrv))

	go ds.Handle(m)

	mockSrv.recvChan <- makeSeek(systemChainID, &ab.SeekInfo{Start: seekSpecified(uint64(ledgerSize - 1)), Stop: seekSpecified(ledgerSize), Behavior: ab.SeekInfo_BLOCK_UNTIL_READY})

	select {
	case deliverReply := <-mockSrv.sendChan:
		assert.Equal(t, cb.Status_SERVICE_UNAVAILABLE, deliverReply.GetStatus(), "Mock support errored")
	case <-time.After(time.Second):
		t.Fatalf("Timed out waiting for error response")
	}
}

func TestCanceledContext(t *testing.T) {
	mm := newMockMultichainManager()
	ms := mm.chains[systemChainID]
	l := ms.ledger
	for i := 1; i < ledgerSize; i++ {
		l.Append(blockledger.CreateNextBlock(l, []*cb.Envelope{{Payload: []byte(fmt.Sprintf("%d", i))}}))
	}

	mockSrv := newMockD()
	defer close(mockSrv.recvChan)
	ds := initializeDeliverHandler(mm, !mutualTLS, false)

	errCh := make(chan error, 1)

	go func() {
		err := ds.Handle(NewDeliverServer(mockSrv, mm.PolicyChecker, sendDeliverResponseProducer(mockSrv)))
		errCh <- err
	}()

	mockSrv.recvChan <- makeSeek(systemChainID, &ab.SeekInfo{Start: seekSpecified(uint64(ledgerSize - 1)), Stop: seekSpecified(ledgerSize), Behavior: ab.SeekInfo_BLOCK_UNTIL_READY})

	select {
	case deliverReply := <-mockSrv.sendChan:
		assert.NotNil(t, deliverReply.GetBlock(), "Expected first block")
	case <-time.After(time.Second):
		t.Fatalf("Timed out waiting to get first block")
	}

	mockSrv.CancelContext()

	select {
	case err := <-errCh:
		assert.Error(t, err)
		assert.Regexp(t, "context canceled", err.Error())
	case <-time.After(time.Second):
		t.Fatalf("Timed out waiting for error response")
	}
}

func TestErroredBlockingSeek(t *testing.T) {
	mm := newMockMultichainManager()
	ms := mm.chains[systemChainID]
	l := ms.ledger
	for i := 1; i < ledgerSize; i++ {
		l.Append(blockledger.CreateNextBlock(l, []*cb.Envelope{{Payload: []byte(fmt.Sprintf("%d", i))}}))
	}
=======
			channelHeaderPayload = nil
			seekInfoPayload = nil
>>>>>>> 95483dd862a768c588c2582e88f6b9da37a47ed7

			envelope = &cb.Envelope{}
			fakeReceiver.RecvReturns(envelope, nil)
			fakeReceiver.RecvReturnsOnCall(1, nil, io.EOF)
		})

<<<<<<< HEAD
	go ds.Handle(NewDeliverServer(mockSrv, mm.PolicyChecker, sendDeliverResponseProducer(mockSrv)))

	mockSrv.recvChan <- makeSeek(systemChainID, &ab.SeekInfo{Start: seekSpecified(uint64(ledgerSize - 1)), Stop: seekSpecified(ledgerSize), Behavior: ab.SeekInfo_BLOCK_UNTIL_READY})

	select {
	case deliverReply := <-mockSrv.sendChan:
		assert.NotNil(t, deliverReply.GetBlock(), "Expected first block")
	case <-time.After(time.Second):
		t.Fatalf("Timed out waiting to get first block")
	}

	close(ms.erroredChan)

	select {
	case deliverReply := <-mockSrv.sendChan:
		assert.Equal(t, cb.Status_SERVICE_UNAVAILABLE, deliverReply.GetStatus(), "Mock support errored")
	case <-time.After(time.Second):
		t.Fatalf("Timed out waiting for error response")
	}
}

func TestSGracefulShutdown(t *testing.T) {
	mockSrv := newMockD()
	mm := newMockMultichainManager()
	ds := initializeDeliverHandler(mm, !mutualTLS, true)

	close(mockSrv.recvChan)
	assert.NoError(t, ds.Handle(NewDeliverServer(mockSrv, mm.PolicyChecker, sendDeliverResponseProducer(mockSrv))), "Expected no error for hangup")
}

func TestReversedSeqSeek(t *testing.T) {
	mockSrv := newMockD()
	defer close(mockSrv.recvChan)

	mm := newMockMultichainManager()
	ds := initializeDeliverHandler(mm, !mutualTLS, true)
	go ds.Handle(NewDeliverServer(mockSrv, mm.PolicyChecker, sendDeliverResponseProducer(mockSrv)))

	specifiedStart := uint64(7)
	specifiedStop := uint64(3)
	mockSrv.recvChan <- makeSeek(systemChainID, &ab.SeekInfo{Start: seekSpecified(specifiedStart), Stop: seekSpecified(specifiedStop), Behavior: ab.SeekInfo_BLOCK_UNTIL_READY})

	select {
	case deliverReply := <-mockSrv.sendChan:
		if deliverReply.GetStatus() != cb.Status_BAD_REQUEST {
			t.Fatalf("Received wrong error on the reply channel")
		}
	case <-time.After(time.Second):
		t.Fatalf("Timed out waiting to get all blocks")
	}
}

func TestBadStreamRecv(t *testing.T) {
	mm := newMockMultichainManager()
	bh := initializeDeliverHandler(mm, !mutualTLS, true)
	assert.Error(t, bh.Handle(NewDeliverServer(&erroneousRecvMockD{}, mm.PolicyChecker, sendDeliverResponseProducer(&erroneousRecvMockD{}))), "Should catch unexpected stream error")
}

func TestBadStreamSend(t *testing.T) {
	m := &erroneousSendMockD{recvVal: makeSeek(systemChainID, &ab.SeekInfo{Start: seekNewest, Stop: seekNewest, Behavior: ab.SeekInfo_BLOCK_UNTIL_READY})}
	mm := newMockMultichainManager()
	ds := initializeDeliverHandler(mm, !mutualTLS, true)
	assert.Error(t, ds.Handle(NewDeliverServer(m, mm.PolicyChecker, sendDeliverResponseProducer(m))), "Should catch unexpected stream error")
}

func TestOldestSeek(t *testing.T) {
	mockSrv := newMockD()
	defer close(mockSrv.recvChan)

	mm := newMockMultichainManager()
	ds := initializeDeliverHandler(mm, !mutualTLS, true)
	go ds.Handle(NewDeliverServer(mockSrv, mm.PolicyChecker, sendDeliverResponseProducer(mockSrv)))

	mockSrv.recvChan <- makeSeek(systemChainID, &ab.SeekInfo{Start: seekOldest, Stop: seekOldest, Behavior: ab.SeekInfo_BLOCK_UNTIL_READY})

	select {
	case deliverReply := <-mockSrv.sendChan:
		assert.NotEqual(t, nil, deliverReply.GetBlock(), "Received an error on the reply channel")
		assert.Equal(t, uint64(0), deliverReply.GetBlock().Header.Number, "Expected only the most recent block")
	case <-time.After(time.Second):
		t.Fatalf("Timed out waiting to get all blocks")
	}
}

func TestNoPayloadSeek(t *testing.T) {
	mockSrv := newMockD()
	defer close(mockSrv.recvChan)
	mm := newMockMultichainManager()
	ds := initializeDeliverHandler(mm, !mutualTLS, true)
	go ds.Handle(NewDeliverServer(mockSrv, mm.PolicyChecker, sendDeliverResponseProducer(mockSrv)))

	mockSrv.recvChan <- &cb.Envelope{Payload: []byte("Foo")}

	select {
	case deliverReply := <-mockSrv.sendChan:
		assert.Equal(t, cb.Status_BAD_REQUEST, deliverReply.GetStatus(), "Received wrong error on the reply channel")
	case <-time.After(time.Second):
		t.Fatalf("Timed out waiting to get all blocks")
	}
}

func TestNilPayloadHeaderSeek(t *testing.T) {
	mockSrv := newMockD()
	defer close(mockSrv.recvChan)
	mm := newMockMultichainManager()
	ds := initializeDeliverHandler(mm, !mutualTLS, true)
	go ds.Handle(NewDeliverServer(mockSrv, mm.PolicyChecker, sendDeliverResponseProducer(mockSrv)))

	mockSrv.recvChan <- &cb.Envelope{Payload: utils.MarshalOrPanic(&cb.Payload{})}

	select {
	case deliverReply := <-mockSrv.sendChan:
		assert.Equal(t, cb.Status_BAD_REQUEST, deliverReply.GetStatus(), "Received wrong error on the reply channel")
	case <-time.After(time.Second):
		t.Fatalf("Timed out waiting to get all blocks")
	}
}

func TestBadChannelHeader(t *testing.T) {
	mockSrv := newMockD()
	defer close(mockSrv.recvChan)

	mm := newMockMultichainManager()
	ds := initializeDeliverHandler(mm, !mutualTLS, true)
	go ds.Handle(NewDeliverServer(mockSrv, mm.PolicyChecker, sendDeliverResponseProducer(mockSrv)))

	mockSrv.recvChan <- &cb.Envelope{Payload: utils.MarshalOrPanic(&cb.Payload{
		Header: &cb.Header{ChannelHeader: []byte("Foo")},
	})}

	select {
	case deliverReply := <-mockSrv.sendChan:
		assert.Equal(t, cb.Status_BAD_REQUEST, deliverReply.GetStatus(), "Received wrong error on the reply channel")
	case <-time.After(time.Second):
		t.Fatalf("Timed out waiting to get all blocks")
	}
}

func TestChainNotFound(t *testing.T) {
	mm := &mockSupportManager{
		chains: make(map[string]*mockSupport),
	}

	mockSrv := newMockD()
	defer close(mockSrv.recvChan)
	ds := initializeDeliverHandler(mm, !mutualTLS, false)

	go ds.Handle(NewDeliverServer(mockSrv, mm.PolicyChecker, sendDeliverResponseProducer(mockSrv)))

	mockSrv.recvChan <- makeSeek(systemChainID, &ab.SeekInfo{Start: seekNewest, Stop: seekNewest, Behavior: ab.SeekInfo_BLOCK_UNTIL_READY})

	select {
	case deliverReply := <-mockSrv.sendChan:
		assert.Equal(t, cb.Status_NOT_FOUND, deliverReply.GetStatus(), "Received wrong error on the reply channel")
	case <-time.After(time.Second):
		t.Fatalf("Timed out waiting to get all blocks")
	}
}

func TestBadSeekInfoPayload(t *testing.T) {
	mockSrv := newMockD()
	defer close(mockSrv.recvChan)
	mm := newMockMultichainManager()
	ds := initializeDeliverHandler(mm, !mutualTLS, true)
	go ds.Handle(NewDeliverServer(mockSrv, mm.PolicyChecker, sendDeliverResponseProducer(mockSrv)))

	mockSrv.recvChan <- &cb.Envelope{
		Payload: utils.MarshalOrPanic(&cb.Payload{
			Header: &cb.Header{
				ChannelHeader: utils.MarshalOrPanic(&cb.ChannelHeader{
					ChannelId: systemChainID,
					Timestamp: util.CreateUtcTimestamp(),
				}),
				SignatureHeader: utils.MarshalOrPanic(&cb.SignatureHeader{}),
			},
			Data: []byte("Foo"),
		}),
	}

	select {
	case deliverReply := <-mockSrv.sendChan:
		assert.Equal(t, cb.Status_BAD_REQUEST, deliverReply.GetStatus(), "Received wrong error on the reply channel")
	case <-time.After(time.Second):
		t.Fatalf("Timed out waiting to get all blocks")
	}
}

func TestMissingSeekPosition(t *testing.T) {
	mockSrv := newMockD()
	defer close(mockSrv.recvChan)
	mm := newMockMultichainManager()
	ds := initializeDeliverHandler(mm, !mutualTLS, true)
	go ds.Handle(NewDeliverServer(mockSrv, mm.PolicyChecker, sendDeliverResponseProducer(mockSrv)))

	mockSrv.recvChan <- &cb.Envelope{
		Payload: utils.MarshalOrPanic(&cb.Payload{
			Header: &cb.Header{
				ChannelHeader: utils.MarshalOrPanic(&cb.ChannelHeader{
					ChannelId: systemChainID,
					Timestamp: util.CreateUtcTimestamp(),
				}),
				SignatureHeader: utils.MarshalOrPanic(&cb.SignatureHeader{}),
			},
			Data: nil,
		}),
	}

	select {
	case deliverReply := <-mockSrv.sendChan:
		assert.Equal(t, cb.Status_BAD_REQUEST, deliverReply.GetStatus(), "Received wrong error on the reply channel")
	case <-time.After(time.Second):
		t.Fatalf("Timed out waiting to get all blocks")
	}
}

func TestNilTimestamp(t *testing.T) {
	mockSrv := newMockD()
	defer close(mockSrv.recvChan)
	mm := newMockMultichainManager()
	ds := initializeDeliverHandler(mm, !mutualTLS, true)
	go ds.Handle(NewDeliverServer(mockSrv, mm.PolicyChecker, sendDeliverResponseProducer(mockSrv)))

	mockSrv.recvChan <- &cb.Envelope{
		Payload: utils.MarshalOrPanic(&cb.Payload{
			Header: &cb.Header{
				ChannelHeader: utils.MarshalOrPanic(&cb.ChannelHeader{
					ChannelId: systemChainID,
				}),
				SignatureHeader: utils.MarshalOrPanic(&cb.SignatureHeader{}),
			},
			Data: nil,
		}),
	}

	select {
	case deliverReply := <-mockSrv.sendChan:
		assert.Equal(t, cb.Status_BAD_REQUEST, deliverReply.GetStatus(), "Received wrong error on the reply channel")
	case <-time.After(time.Second):
		t.Fatalf("Timed out waiting to get all blocks")
	}
}

func TestTimestampOutOfTimeWindow(t *testing.T) {
	mockSrv := newMockD()
	defer close(mockSrv.recvChan)
	mm := newMockMultichainManager()
	ds := initializeDeliverHandler(mm, !mutualTLS, true)
	go ds.Handle(NewDeliverServer(mockSrv, mm.PolicyChecker, sendDeliverResponseProducer(mockSrv)))

	mockSrv.recvChan <- &cb.Envelope{
		Payload: utils.MarshalOrPanic(&cb.Payload{
			Header: &cb.Header{
				ChannelHeader: utils.MarshalOrPanic(&cb.ChannelHeader{
					ChannelId: systemChainID,
					Timestamp: &timestamp.Timestamp{Seconds: 0},
				}),
				SignatureHeader: utils.MarshalOrPanic(&cb.SignatureHeader{}),
			},
			Data: nil,
		}),
	}

	select {
	case deliverReply := <-mockSrv.sendChan:
		assert.Equal(t, cb.Status_BAD_REQUEST, deliverReply.GetStatus(), "Received wrong error on the reply channel")
	case <-time.After(time.Second):
		t.Fatalf("Timed out waiting to get all blocks")
	}
}

func TestSeekWithMutualTLS(t *testing.T) {
	mockSrv := newMockD()
	mockSrv.StreamContext = nil
	defer close(mockSrv.recvChan)
	mm := newMockMultichainManager()
	ds := initializeDeliverHandler(mm, mutualTLS, true)
	go ds.Handle(NewDeliverServer(mockSrv, mm.PolicyChecker, sendDeliverResponseProducer(mockSrv)))

	mockSrv.recvChan <- makeSeekWithTLSCertHash(systemChainID, &ab.SeekInfo{Start: seekNewest, Stop: seekNewest, Behavior: ab.SeekInfo_BLOCK_UNTIL_READY}, testCert)

	select {
	case deliverReply := <-mockSrv.sendChan:
		if deliverReply.GetBlock() == nil {
			t.Fatalf("Received an error on the reply channel")
		}
		if deliverReply.GetBlock().Header.Number != uint64(ledgerSize-1) {
			t.Fatalf("Expected only the most recent block")
		}
	case <-time.After(time.Second):
		t.Fatalf("Timed out waiting to get all blocks")
	}
}

func TestSeekWithMutualTLS_wrongTLSCert(t *testing.T) {
	mockSrv := newMockD()
	defer close(mockSrv.recvChan)
	mm := newMockMultichainManager()
	ds := initializeDeliverHandler(mm, mutualTLS, true)
	go ds.Handle(NewDeliverServer(mockSrv, mm.PolicyChecker, sendDeliverResponseProducer(mockSrv)))
	wrongCert := &x509.Certificate{
		Raw: []byte("wrong"),
	}
	mockSrv.recvChan <- makeSeekWithTLSCertHash(systemChainID, &ab.SeekInfo{Start: seekNewest, Stop: seekNewest, Behavior: ab.SeekInfo_BLOCK_UNTIL_READY}, wrongCert)

	select {
	case deliverReply := <-mockSrv.sendChan:
		assert.Equal(t, cb.Status_BAD_REQUEST, deliverReply.GetStatus(), "Received wrong error on the reply channel")
	case <-time.After(time.Second):
		t.Fatalf("Timed out waiting to get all blocks")
	}
}

func TestSeekWithMutualTLS_noTLSCert(t *testing.T) {
	mockSrv := newMockD()
	defer close(mockSrv.recvChan)
	mm := newMockMultichainManager()
	ds := initializeDeliverHandler(mm, mutualTLS, true)
	go ds.Handle(NewDeliverServer(mockSrv, mm.PolicyChecker, sendDeliverResponseProducer(mockSrv)))

	mockSrv.recvChan <- makeSeek(systemChainID, &ab.SeekInfo{Start: seekNewest, Stop: seekNewest, Behavior: ab.SeekInfo_BLOCK_UNTIL_READY})

	select {
	case deliverReply := <-mockSrv.sendChan:
		assert.Equal(t, cb.Status_BAD_REQUEST, deliverReply.GetStatus(), "Received wrong error on the reply channel")
	case <-time.After(time.Second):
		t.Fatalf("Timed out waiting to get all blocks")
	}
}
=======
		JustBeforeEach(func() {
			if channelHeaderPayload == nil {
				channelHeaderPayload = utils.MarshalOrPanic(channelHeader)
			}
			if seekInfoPayload == nil {
				seekInfoPayload = utils.MarshalOrPanic(seekInfo)
			}
			if envelope.Payload == nil {
				payload := &cb.Payload{
					Header: &cb.Header{
						ChannelHeader:   channelHeaderPayload,
						SignatureHeader: utils.MarshalOrPanic(&cb.SignatureHeader{}),
					},
					Data: seekInfoPayload,
				}
				envelope.Payload = utils.MarshalOrPanic(payload)
			}
		})

		It("validates the channel header with the binding inspector", func() {
			err := handler.Handle(context.Background(), server)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeInspector.InspectCallCount()).To(Equal(1))
			ctx, header := fakeInspector.InspectArgsForCall(0)
			Expect(ctx).To(Equal(context.Background()))
			Expect(header).To(Equal(channelHeader))
		})

		Context("when channel header validation fails", func() {
			BeforeEach(func() {
				fakeInspector.InspectReturns(errors.New("bad-header-thingy"))
			})

			It("sends a bad request message", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeResponseSender.SendStatusResponseCallCount()).To(Equal(1))
				resp := fakeResponseSender.SendStatusResponseArgsForCall(0)
				Expect(resp).To(Equal(cb.Status_BAD_REQUEST))
			})
		})

		It("gets the chain from the chain manager", func() {
			err := handler.Handle(context.Background(), server)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeChainManager.GetChainCallCount()).To(Equal(1))
			chid := fakeChainManager.GetChainArgsForCall(0)
			Expect(chid).To(Equal("chain-id"))
		})

		It("receives messages until io.EOF is returned", func() {
			err := handler.Handle(context.Background(), server)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeReceiver.RecvCallCount()).To(Equal(2))
		})

		It("evaluates access control", func() {
			err := handler.Handle(context.Background(), server)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakePolicyChecker.CheckPolicyCallCount()).To(BeNumerically(">=", 1))
			e, cid := fakePolicyChecker.CheckPolicyArgsForCall(0)
			Expect(e).To(Equal(envelope))
			Expect(cid).To(Equal("chain-id"))
		})

		It("gets a block iterator from the starting block", func() {
			err := handler.Handle(context.Background(), server)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeBlockReader.IteratorCallCount()).To(Equal(1))
			startPosition := fakeBlockReader.IteratorArgsForCall(0)
			Expect(startPosition).To(Equal(seekInfo.Start))
		})

		Context("when multiple blocks are requested", func() {
			BeforeEach(func() {
				fakeBlockIterator.NextStub = func() (*cb.Block, cb.Status) {
					blk := &cb.Block{
						Header: &cb.BlockHeader{Number: 994 + uint64(fakeBlockIterator.NextCallCount())},
					}
					return blk, cb.Status_SUCCESS
				}
				seekInfo = &ab.SeekInfo{
					Start: &ab.SeekPosition{
						Type: &ab.SeekPosition_Specified{Specified: &ab.SeekSpecified{Number: 995}},
					},
					Stop: seekNewest,
				}
			})

			It("sends all requested blocks", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeResponseSender.SendBlockResponseCallCount()).To(Equal(5))
				for i := 0; i < 5; i++ {
					b := fakeResponseSender.SendBlockResponseArgsForCall(i)
					Expect(b).To(Equal(&cb.Block{
						Header: &cb.BlockHeader{Number: 995 + uint64(i)},
					}))
				}
			})
		})

		Context("when seek info is configured to stop at the oldest block", func() {
			BeforeEach(func() {
				seekInfo = &ab.SeekInfo{Start: &ab.SeekPosition{}, Stop: seekOldest}
			})

			It("sends only the first block returned by the iterator", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeBlockReader.IteratorCallCount()).To(Equal(1))
				start := fakeBlockReader.IteratorArgsForCall(0)
				Expect(start).To(Equal(&ab.SeekPosition{}))
				Expect(fakeBlockIterator.NextCallCount()).To(Equal(1))

				Expect(fakeResponseSender.SendBlockResponseCallCount()).To(Equal(1))
				b := fakeResponseSender.SendBlockResponseArgsForCall(0)
				Expect(b).To(Equal(&cb.Block{
					Header: &cb.BlockHeader{Number: 100},
				}))
			})
		})

		Context("when seek info is configured to stop at the newest block", func() {
			BeforeEach(func() {
				seekInfo = &ab.SeekInfo{Start: &ab.SeekPosition{}, Stop: seekNewest}

				fakeBlockReader.HeightReturns(3)
				fakeBlockIterator.NextStub = func() (*cb.Block, cb.Status) {
					blk := &cb.Block{
						Header: &cb.BlockHeader{Number: uint64(fakeBlockIterator.NextCallCount())},
					}
					return blk, cb.Status_SUCCESS
				}
			})

			It("sends blocks until the iterator reaches the reader height", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeBlockReader.IteratorCallCount()).To(Equal(1))
				start := fakeBlockReader.IteratorArgsForCall(0)
				Expect(start).To(Equal(&ab.SeekPosition{}))

				Expect(fakeBlockIterator.NextCallCount()).To(Equal(2))
				Expect(fakeResponseSender.SendBlockResponseCallCount()).To(Equal(2))
				for i := 0; i < fakeResponseSender.SendBlockResponseCallCount(); i++ {
					b := fakeResponseSender.SendBlockResponseArgsForCall(i)
					Expect(b).To(Equal(&cb.Block{
						Header: &cb.BlockHeader{Number: uint64(i + 1)},
					}))
				}
			})
		})

		Context("when sending the block fails", func() {
			BeforeEach(func() {
				fakeResponseSender.SendBlockResponseReturns(errors.New("send-fails"))
			})

			It("returns the error", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).To(MatchError("send-fails"))
			})
		})

		It("sends a success response", func() {
			err := handler.Handle(context.Background(), server)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeResponseSender.SendStatusResponseCallCount()).To(Equal(1))
			resp := fakeResponseSender.SendStatusResponseArgsForCall(0)
			Expect(resp).To(Equal(cb.Status_SUCCESS))
		})

		Context("when sending the success status fails", func() {
			BeforeEach(func() {
				fakeResponseSender.SendStatusResponseReturns(errors.New("send-success-fails"))
			})

			It("returns the error", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).To(MatchError("send-success-fails"))
			})
		})

		Context("when receive fails", func() {
			BeforeEach(func() {
				fakeReceiver.RecvReturns(nil, errors.New("oh bother"))
			})

			It("returns the error", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).To(MatchError("oh bother"))
			})
		})

		Context("when unmarshaling the payload fails", func() {
			BeforeEach(func() {
				envelope.Payload = []byte("completely-bogus-data")
			})

			It("sends a bad request message", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeResponseSender.SendStatusResponseCallCount()).To(Equal(1))
				resp := fakeResponseSender.SendStatusResponseArgsForCall(0)
				Expect(resp).To(Equal(cb.Status_BAD_REQUEST))
			})
		})

		Context("when the payload header is nil", func() {
			BeforeEach(func() {
				envelope.Payload = utils.MarshalOrPanic(&cb.Payload{
					Header: nil,
				})
			})

			It("sends a bad request message", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeResponseSender.SendStatusResponseCallCount()).To(Equal(1))
				resp := fakeResponseSender.SendStatusResponseArgsForCall(0)
				Expect(resp).To(Equal(cb.Status_BAD_REQUEST))
			})
		})

		Context("when unmarshaling the channel header fails", func() {
			BeforeEach(func() {
				channelHeaderPayload = []byte("complete-nonsense")
			})

			It("sends a bad request message", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeResponseSender.SendStatusResponseCallCount()).To(Equal(1))
				resp := fakeResponseSender.SendStatusResponseArgsForCall(0)
				Expect(resp).To(Equal(cb.Status_BAD_REQUEST))
			})
		})

		Context("when the channel header timestamp is nil", func() {
			BeforeEach(func() {
				channelHeaderPayload = utils.MarshalOrPanic(&cb.ChannelHeader{
					Timestamp: nil,
				})
			})

			It("sends a bad request message", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeResponseSender.SendStatusResponseCallCount()).To(Equal(1))
				resp := fakeResponseSender.SendStatusResponseArgsForCall(0)
				Expect(resp).To(Equal(cb.Status_BAD_REQUEST))
			})
		})

		Context("when the channel header timestamp is out of the time window", func() {
			BeforeEach(func() {
				channelHeaderPayload = utils.MarshalOrPanic(&cb.ChannelHeader{
					Timestamp: &timestamp.Timestamp{},
				})
			})

			It("sends status bad request", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeResponseSender.SendStatusResponseCallCount()).To(Equal(1))
				resp := fakeResponseSender.SendStatusResponseArgsForCall(0)
				Expect(resp).To(Equal(cb.Status_BAD_REQUEST))
			})
		})

		Context("when the channel is not found", func() {
			BeforeEach(func() {
				fakeChainManager.GetChainReturns(nil, false)
			})

			It("sends status not found", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeResponseSender.SendStatusResponseCallCount()).To(Equal(1))
				resp := fakeResponseSender.SendStatusResponseArgsForCall(0)
				Expect(resp).To(Equal(cb.Status_NOT_FOUND))
			})
		})

		Context("when the client disconnects before reading from the chain", func() {
			var (
				ctx    context.Context
				cancel func()
				done   chan struct{}
			)

			BeforeEach(func() {
				done = make(chan struct{})
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
				fakeBlockIterator.NextStub = func() (*cb.Block, cb.Status) {
					<-done
					return nil, cb.Status_BAD_REQUEST
				}
			})

			AfterEach(func() {
				close(done)
			})

			It("aborts the deliver stream", func() {
				err := handler.Handle(ctx, server)
				Expect(err).To(MatchError("context finished before block retrieved: context canceled"))
			})
		})

		Context("when the chain errors before reading from the chain", func() {
			BeforeEach(func() {
				close(errCh)
			})

			It("sends status service unavailable", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeChain.ReaderCallCount()).To(Equal(0))
				Expect(fakeResponseSender.SendStatusResponseCallCount()).To(Equal(1))
				resp := fakeResponseSender.SendStatusResponseArgsForCall(0)
				Expect(resp).To(Equal(cb.Status_SERVICE_UNAVAILABLE))
			})
		})

		Context("when the chain errors while reading from the chain", func() {
			BeforeEach(func() {
				fakeChain.ReaderStub = func() blockledger.Reader {
					close(errCh)
					return fakeBlockReader
				}
			})

			It("sends status service unavailable", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeChain.ReaderCallCount()).To(Equal(1))
				Expect(fakeResponseSender.SendStatusResponseCallCount()).To(Equal(1))
				resp := fakeResponseSender.SendStatusResponseArgsForCall(0)
				Expect(resp).To(Equal(cb.Status_SERVICE_UNAVAILABLE))
			})
		})

		Context("when the access evaluation fails", func() {
			BeforeEach(func() {
				fakePolicyChecker.CheckPolicyReturns(errors.New("no-access-for-you"))
			})

			It("sends status not found", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeResponseSender.SendStatusResponseCallCount()).To(Equal(1))
				resp := fakeResponseSender.SendStatusResponseArgsForCall(0)
				Expect(resp).To(Equal(cb.Status_FORBIDDEN))
			})
		})

		Context("when the access expires", func() {
			BeforeEach(func() {
				fakeChain.SequenceStub = func() uint64 {
					return uint64(fakeChain.SequenceCallCount())
				}
				fakePolicyChecker.CheckPolicyReturnsOnCall(1, errors.New("no-access-for-you"))
			})

			It("sends status not found", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeResponseSender.SendStatusResponseCallCount()).To(Equal(1))
				resp := fakeResponseSender.SendStatusResponseArgsForCall(0)
				Expect(resp).To(Equal(cb.Status_FORBIDDEN))

				Expect(fakePolicyChecker.CheckPolicyCallCount()).To(Equal(2))
			})
		})

		Context("when unmarshaling seek info fails", func() {
			BeforeEach(func() {
				seekInfoPayload = []byte("complete-nonsense")
			})

			It("sends status bad request", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeResponseSender.SendStatusResponseCallCount()).To(Equal(1))
				resp := fakeResponseSender.SendStatusResponseArgsForCall(0)
				Expect(resp).To(Equal(cb.Status_BAD_REQUEST))
			})
		})

		Context("when seek start and stop are nil", func() {
			BeforeEach(func() {
				seekInfo = &ab.SeekInfo{Start: nil, Stop: nil}
			})

			It("sends status bad request", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeResponseSender.SendStatusResponseCallCount()).To(Equal(1))
				resp := fakeResponseSender.SendStatusResponseArgsForCall(0)
				Expect(resp).To(Equal(cb.Status_BAD_REQUEST))
			})
		})

		Context("when seek info start number is greater than stop number", func() {
			BeforeEach(func() {
				seekInfo = &ab.SeekInfo{
					Start: seekNewest,
					Stop: &ab.SeekPosition{
						Type: &ab.SeekPosition_Specified{Specified: &ab.SeekSpecified{Number: 99}},
					},
				}
			})

			It("sends status bad request", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeResponseSender.SendStatusResponseCallCount()).To(Equal(1))
				resp := fakeResponseSender.SendStatusResponseArgsForCall(0)
				Expect(resp).To(Equal(cb.Status_BAD_REQUEST))
			})
		})

		Context("when fail if not ready is set and the next block is unavailable", func() {
			BeforeEach(func() {
				fakeBlockReader.HeightReturns(1000)
				fakeBlockReader.IteratorReturns(fakeBlockIterator, 1000)

				seekInfo = &ab.SeekInfo{
					Behavior: ab.SeekInfo_FAIL_IF_NOT_READY,
					Start: &ab.SeekPosition{
						Type: &ab.SeekPosition_Specified{Specified: &ab.SeekSpecified{Number: 1002}},
					},
					Stop: &ab.SeekPosition{
						Type: &ab.SeekPosition_Specified{Specified: &ab.SeekSpecified{Number: 1003}},
					},
				}
			})

			It("sends status not found", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeBlockIterator.NextCallCount()).To(Equal(0))
				Expect(fakeResponseSender.SendStatusResponseCallCount()).To(Equal(1))
				resp := fakeResponseSender.SendStatusResponseArgsForCall(0)
				Expect(resp).To(Equal(cb.Status_NOT_FOUND))
			})
		})

		Context("when next block status does not indicate success", func() {
			BeforeEach(func() {
				fakeBlockIterator.NextReturns(nil, cb.Status_UNKNOWN)
			})

			It("forwards the status response", func() {
				err := handler.Handle(context.Background(), server)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeResponseSender.SendStatusResponseCallCount()).To(Equal(1))
				resp := fakeResponseSender.SendStatusResponseArgsForCall(0)
				Expect(resp).To(Equal(cb.Status_UNKNOWN))
			})
		})
	})
})
>>>>>>> 95483dd862a768c588c2582e88f6b9da37a47ed7
