package fabric

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/fabric-creed/fabric-hub/pkg/fabric/util"
	"github.com/fabric-creed/fabric-protos-go/common"
	"github.com/fabric-creed/fabric-protos-go/ledger/rwset"
	"github.com/fabric-creed/fabric-protos-go/ledger/rwset/kvrwset"
	"github.com/fabric-creed/fabric-protos-go/msp"
	"github.com/fabric-creed/fabric-protos-go/peer"
	"github.com/fabric-creed/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

type BlockchainInfo struct {
	Height            uint64 `json:"height,omitempty"`
	CurrentBlockHash  string `json:"current_block_hash,omitempty"`
	PreviousBlockHash string `json:"previous_block_hash,omitempty"`
}

func UnmarshalBlockchainInfo(raw []byte) (*BlockchainInfo, error) {
	in := &common.BlockchainInfo{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodeBlockchainInfo(in)
}

func DecodeBlockchainInfo(in *common.BlockchainInfo) (*BlockchainInfo, error) {
	if in == nil {
		return nil, nil
	}
	out := &BlockchainInfo{
		Height:            in.Height,
		CurrentBlockHash:  hex.EncodeToString(in.CurrentBlockHash),
		PreviousBlockHash: hex.EncodeToString(in.PreviousBlockHash),
	}
	return out, nil
}

type Block struct {
	OriginData []byte         `json:"origin_data,omitempty"`
	BlockHash  string         `json:"block_hash,omitempty"`
	BlockTime  int64          `json:"block_time,omitempty"`
	Header     *BlockHeader   `json:"header,omitempty"`
	Data       *BlockData     `json:"data,omitempty"`
	Metadata   *BlockMetadata `json:"metadata,omitempty"`
}

func DecodeBlock(in *common.Block, isGM bool) (*Block, error) {
	if in == nil {
		return nil, nil
	}
	out := &Block{}
	outOriginData, err := proto.Marshal(in)
	if err != nil {
		logrus.Errorf("failed to marshal in(%+v): %v", in, err)
	} else {
		out.OriginData = outOriginData
	}
	outBlockHash := util.BlockHeaderHash(in.Header, isGM)
	out.BlockHash = hex.EncodeToString(outBlockHash)
	outHeader, err := DecodeBlockHeader(in.Header)
	if err != nil {
		logrus.Errorf("failed to decode in.Header(%+v): %v", in.Header, err)
	} else {
		out.Header = outHeader
	}
	outData, err := DecodeBlockData(in.Data)
	if err != nil {
		logrus.Errorf("failed to decode in.Data(%+v): %v", in.Data, err)
	} else {
		out.Data = outData
	}
	outMetadata, err := DecodeBlockMetadata(in.Metadata)
	if err != nil {
		logrus.Errorf("failed to decode in.Metadata(%+v): %v", in.Metadata, err)
	} else {
		out.Metadata = outMetadata
	}
	out.FillBlockTime()
	return out, nil
}

func (b *Block) FillBlockTime() {
	if b.Data == nil {
		return
	}
	if len(b.Data.Data) == 0 {
		return
	}
	envelope := b.Data.Data[0]
	if envelope.Payload == nil {
		return
	}
	payload := envelope.Payload
	if payload.Header == nil {
		return
	}
	header := payload.Header
	if header.ChannelHeader == nil {
		return
	}
	channelHeader := header.ChannelHeader
	b.BlockTime = channelHeader.Timestamp
}

type BlockHeader struct {
	Number       uint64 `json:"number,omitempty"`
	PreviousHash string `json:"previous_hash,omitempty"`
	DataHash     string `json:"data_hash,omitempty"`
}

func DecodeBlockHeader(in *common.BlockHeader) (*BlockHeader, error) {
	if in == nil {
		return nil, nil
	}
	out := &BlockHeader{
		Number:       in.Number,
		PreviousHash: hex.EncodeToString(in.PreviousHash),
		DataHash:     hex.EncodeToString(in.DataHash),
	}
	return out, nil
}

type BlockData struct {
	Data []*Envelope `json:"data,omitempty"`
}

func DecodeBlockData(in *common.BlockData) (*BlockData, error) {
	if in == nil {
		return nil, nil
	}
	out := &BlockData{}
	for _, inData := range in.Data {
		outData, err := UnmarshalEnvelope(inData)
		if err != nil {
			logrus.Errorf("failed to unmarshal inData(%s): %v", inData, err)
		} else {
			out.Data = append(out.Data, outData)
		}
	}
	return out, nil
}

type BlockMetadata struct {
	Metadata [][]byte `json:"metadata,omitempty"`
}

func DecodeBlockMetadata(in *common.BlockMetadata) (*BlockMetadata, error) {
	if in == nil {
		return nil, nil
	}
	out := &BlockMetadata{
		Metadata: in.Metadata,
	}
	return out, nil
}

type ProcessedTransaction struct {
	TransactionEnvelope *Envelope `json:"transaction_envelope,omitempty"`
	ValidationCode      int32     `json:"validation_code,omitempty"`
}

func DecodeProcessedTransaction(in *peer.ProcessedTransaction) (*ProcessedTransaction, error) {
	if in == nil {
		return nil, nil
	}
	out := &ProcessedTransaction{
		ValidationCode: in.ValidationCode,
	}
	outTransactionEnvelope, err := DecodeEnvelope(in.TransactionEnvelope)
	if err != nil {
		logrus.Errorf("failed to decode in.TransactionEnvelope(%+v): %v", in.TransactionEnvelope, err)
	} else {
		out.TransactionEnvelope = outTransactionEnvelope
	}
	return out, nil
}

type Envelope struct {
	OriginData []byte   `json:"origin_data,omitempty"`
	Payload    *Payload `json:"payload,omitempty"`
	Signature  string   `json:"signature,omitempty"`
}

func UnmarshalEnvelope(raw []byte) (*Envelope, error) {
	in := &common.Envelope{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodeEnvelope(in)
}

func DecodeEnvelope(in *common.Envelope) (*Envelope, error) {
	if in == nil {
		return nil, nil
	}
	out := &Envelope{
		Signature: hex.EncodeToString(in.Signature),
	}
	outOriginData, err := proto.Marshal(in)
	if err != nil {
		logrus.Errorf("failed to marshal in(%+v): %v", in, err)
	} else {
		out.OriginData = outOriginData
	}
	outPayload, err := UnmarshalPayload(in.Payload)
	if err != nil {
		logrus.Errorf("failed to unmarshal in.Payload(%s): %v", in.Payload, err)
	} else {
		out.Payload = outPayload
	}
	return out, nil
}

type ConfigEnvelope struct {
	Config *Config `json:"config,omitempty"`
	// LastUpdate *Envelope `json:"last_update,omitempty"`
	LastUpdate *Envelope `json:"-"`
}

func UnmarshalConfigEnvelope(raw []byte) (*ConfigEnvelope, error) {
	in := &common.ConfigEnvelope{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodeConfigEnvelope(in)
}

func DecodeConfigEnvelope(in *common.ConfigEnvelope) (*ConfigEnvelope, error) {
	if in == nil {
		return nil, nil
	}
	out := &ConfigEnvelope{}
	outConfig, err := DecodeConfig(in.Config)
	if err != nil {
		logrus.Errorf("failed to decode in.Config(%+v): %v", in.Config, err)
	} else {
		out.Config = outConfig
	}
	outLastUpdate, err := DecodeEnvelope(in.LastUpdate)
	if err != nil {
		logrus.Errorf("failed to decode in.LastUpdate(%+v): %v", in.LastUpdate, err)
	} else {
		out.LastUpdate = outLastUpdate
	}
	return out, nil
}

type ConfigUpdateEnvelope struct {
	ConfigUpdate []byte             `json:"config_update,omitempty"`
	Signatures   []*ConfigSignature `json:"signatures,omitempty"`
}

type ConfigSignature struct {
	SignatureHeader *SignatureHeader `json:"signature_header,omitempty"`
	Signature       []byte           `json:"signature,omitempty"`
}

func DecodeConfigSignature(in *common.ConfigSignature) (*ConfigSignature, error) {
	if in == nil {
		return nil, nil
	}
	out := &ConfigSignature{
		Signature: in.Signature,
	}
	outSignatureHeader, err := UnmarshalSignatureHeader(in.SignatureHeader)
	if err != nil {
		logrus.Errorf("failed to unmarshal in.SignatureHeader(%s): %v", in.SignatureHeader, err)
	} else {
		out.SignatureHeader = outSignatureHeader
	}
	return out, nil
}

type Config struct {
	Sequence     uint64       `json:"sequence,omitempty"`
	ChannelGroup *ConfigGroup `json:"channel_group,omitempty"`
}

func DecodeConfig(in *common.Config) (*Config, error) {
	if in == nil {
		return nil, nil
	}
	out := &Config{
		Sequence: in.Sequence,
	}
	outChannelGroup, err := DecodeConfigGroup(in.ChannelGroup)
	if err != nil {
		logrus.Errorf("failed to decode in.ChannelGroup(%+v): %v", in.ChannelGroup, err)
	} else {
		out.ChannelGroup = outChannelGroup
	}
	return out, nil
}

type ConfigGroup struct {
	Version uint64 `json:"version,omitempty"`
	// Groups    map[string]*ConfigGroup  `json:"groups,omitempty"`
	Groups    map[string]*ConfigGroup  `json:"-"`
	Values    map[string]*ConfigValue  `json:"values,omitempty"`
	Policies  map[string]*ConfigPolicy `json:"policies,omitempty"`
	ModPolicy string                   `json:"mod_policy,omitempty"`
}

func DecodeConfigGroup(in *common.ConfigGroup) (*ConfigGroup, error) {
	if in == nil {
		return nil, nil
	}
	out := &ConfigGroup{
		Version:   in.Version,
		ModPolicy: in.ModPolicy,
		Groups:    make(map[string]*ConfigGroup),
		Values:    make(map[string]*ConfigValue),
		Policies:  make(map[string]*ConfigPolicy),
	}
	for key, inGroup := range in.Groups {
		outGroup, err := DecodeConfigGroup(inGroup)
		if err != nil {
			logrus.Errorf("failed to decode inGroup(%+v): %v", inGroup, err)
		} else {
			out.Groups[key] = outGroup
		}
	}
	for key, inValue := range in.Values {
		outValue, err := DecodeConfigValue(inValue)
		if err != nil {
			logrus.Errorf("failed to decode inValue(%+v): %v", inValue, err)
		} else {
			out.Values[key] = outValue
		}
	}
	for key, inPolicy := range in.Policies {
		outPolicy, err := DecodeConfigPolicy(inPolicy)
		if err != nil {
			logrus.Errorf("failed to decode inPolicy(%+v): %v", inPolicy, err)
		} else {
			out.Policies[key] = outPolicy
		}
	}
	return out, nil
}

type ConfigValue struct {
	Version   uint64 `json:"version,omitempty"`
	Value     []byte `json:"value,omitempty"`
	ModPolicy string `json:"mod_policy,omitempty"`
}

func DecodeConfigValue(in *common.ConfigValue) (*ConfigValue, error) {
	if in == nil {
		return nil, nil
	}
	out := &ConfigValue{
		Version:   in.Version,
		Value:     in.Value,
		ModPolicy: in.ModPolicy,
	}
	return out, nil
}

type ConfigPolicy struct {
	Version   uint64  `json:"version,omitempty"`
	Policy    *Policy `json:"policy,omitempty"`
	ModPolicy string  `json:"mod_policy,omitempty"`
}

func DecodeConfigPolicy(in *common.ConfigPolicy) (*ConfigPolicy, error) {
	if in == nil {
		return nil, nil
	}
	out := &ConfigPolicy{
		Version:   in.Version,
		ModPolicy: in.ModPolicy,
	}
	outPolicy, err := DecodePolicy(in.Policy)
	if err != nil {
		logrus.Errorf("failed to decode in.Policy(%+v): %v", in.Policy, err)
	} else {
		out.Policy = outPolicy
	}
	return out, nil
}

type Policy struct {
	Type  int32  `json:"type,omitempty"`
	Value []byte `json:"value,omitempty"`
}

func DecodePolicy(in *common.Policy) (*Policy, error) {
	if in == nil {
		return nil, nil
	}
	out := &Policy{
		Type:  in.Type,
		Value: in.Value,
	}
	return out, nil
}

type Payload struct {
	Header      *Header      `json:"header,omitempty"`
	Transaction *Transaction `json:"transaction,omitempty"`
	Unsupported []byte       `json:"unsupported,omitempty"`
}

func UnmarshalPayload(raw []byte) (*Payload, error) {
	in := &common.Payload{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodePayload(in)
}

func DecodePayload(in *common.Payload) (*Payload, error) {
	if in == nil {
		return nil, nil
	}
	out := &Payload{}
	outHeader, err := DecodeHeader(in.Header)
	if err != nil {
		logrus.Errorf("failed to decode in.Header(%+v): %v", in.Header, err)
	} else {
		out.Header = outHeader
	}
	switch common.HeaderType(outHeader.ChannelHeader.Type) {
	case common.HeaderType_ENDORSER_TRANSACTION:
		outData, err := UnmarshalTransaction(in.Data)
		if err != nil {
			logrus.Errorf("failed to unmarshal in.Data(%s): %v", in.Data, err)
		} else {
			out.Transaction = outData
		}
	default:
		out.Unsupported = in.Data
	}
	return out, nil
}

type Header struct {
	ChannelHeader   *ChannelHeader   `json:"channel_header,omitempty"`
	SignatureHeader *SignatureHeader `json:"signature_header,omitempty"`
}

func DecodeHeader(in *common.Header) (*Header, error) {
	if in == nil {
		return nil, nil
	}
	out := &Header{}
	outChannelHeader, err := UnmarshalChannelHeader(in.ChannelHeader)
	if err != nil {
		logrus.Errorf("failed to unmarshal in.ChannelHeader(%s): %v", in.ChannelHeader, err)
	} else {
		out.ChannelHeader = outChannelHeader
	}
	outSignatureHeader, err := UnmarshalSignatureHeader(in.SignatureHeader)
	if err != nil {
		logrus.Errorf("failed to unmarshal in.SignatureHeader(%s): %v", in.SignatureHeader, err)
	} else {
		out.SignatureHeader = outSignatureHeader
	}
	return out, nil
}

type ChannelHeader struct {
	Type                     int32                     `json:"type,omitempty"`
	Version                  int32                     `json:"version,omitempty"`
	Timestamp                int64                     `json:"timestamp,omitempty"`
	ChannelId                string                    `json:"channel_id,omitempty"`
	TxId                     string                    `json:"tx_id,omitempty"`
	Epoch                    uint64                    `json:"epoch,omitempty"`
	TlsCertHash              string                    `json:"tls_cert_hash,omitempty"`
	ChaincodeHeaderExtension *ChaincodeHeaderExtension `json:"chaincode_header_extension,omitempty"`
}

func UnmarshalChannelHeader(raw []byte) (*ChannelHeader, error) {
	in := &common.ChannelHeader{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodeChannelHeader(in)
}

func DecodeChannelHeader(in *common.ChannelHeader) (*ChannelHeader, error) {
	if in == nil {
		return nil, nil
	}
	out := &ChannelHeader{
		Type:        in.Type,
		Version:     in.Version,
		Timestamp:   in.Timestamp.Seconds,
		ChannelId:   in.ChannelId,
		TxId:        in.TxId,
		Epoch:       in.Epoch,
		TlsCertHash: hex.EncodeToString(in.TlsCertHash),
	}
	outChaincodeHeaderExtension, err := UnmarshalChaincodeHeaderExtension(in.Extension)
	if err != nil {
		logrus.Errorf("failed to unmarshal in.Extension(%s): %v", in.Extension, err)
	} else {
		out.ChaincodeHeaderExtension = outChaincodeHeaderExtension
	}
	return out, nil
}

type ChaincodeHeaderExtension struct {
	ChaincodeId *ChaincodeID `json:"chaincode_id,omitempty"`
}

func UnmarshalChaincodeHeaderExtension(raw []byte) (*ChaincodeHeaderExtension, error) {
	in := &peer.ChaincodeHeaderExtension{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodeChaincodeHeaderExtension(in)
}

func DecodeChaincodeHeaderExtension(in *peer.ChaincodeHeaderExtension) (*ChaincodeHeaderExtension, error) {
	if in == nil {
		return nil, nil
	}
	out := &ChaincodeHeaderExtension{}
	outChaincodeId, err := DecodeChaincodeID(in.ChaincodeId)
	if err != nil {
		logrus.Errorf("failed to decode in.ChaincodeId(%+v): %v", in.ChaincodeId, err)
	} else {
		out.ChaincodeId = outChaincodeId
	}
	return out, nil
}

type SignatureHeader struct {
	Creator *SerializedIdentity `json:"creator,omitempty"`
	Nonce   []byte              `json:"nonce,omitempty"`
}

func UnmarshalSignatureHeader(raw []byte) (*SignatureHeader, error) {
	in := &common.SignatureHeader{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodeSignatureHeader(in)
}

func DecodeSignatureHeader(in *common.SignatureHeader) (*SignatureHeader, error) {
	if in == nil {
		return nil, nil
	}
	out := &SignatureHeader{
		Nonce: in.Nonce,
	}
	outCreator, err := UnmarshalSerializedIdentity(in.Creator)
	if err != nil {
		logrus.Errorf("failed to unmarshal in.Creator(%s): %v", in.Creator, err)
	} else {
		out.Creator = outCreator
	}
	return out, nil
}

type SerializedIdentity struct {
	Mspid   string `json:"mspid,omitempty"`
	IdBytes string `json:"id_bytes,omitempty"`
}

func UnmarshalSerializedIdentity(raw []byte) (*SerializedIdentity, error) {
	in := &msp.SerializedIdentity{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodeSerializedIdentity(in)
}

func DecodeSerializedIdentity(in *msp.SerializedIdentity) (*SerializedIdentity, error) {
	if in == nil {
		return nil, nil
	}
	out := &SerializedIdentity{
		Mspid:   in.Mspid,
		IdBytes: string(in.IdBytes),
	}
	return out, nil
}

type Transaction struct {
	Actions []*TransactionAction `json:"actions,omitempty"`
}

func UnmarshalTransaction(raw []byte) (*Transaction, error) {
	in := &peer.Transaction{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodeTransaction(in)
}

func DecodeTransaction(in *peer.Transaction) (*Transaction, error) {
	if in == nil {
		return nil, nil
	}
	out := &Transaction{}
	for _, inAction := range in.Actions {
		outAction, err := DecodeTransactionAction(inAction)
		if err != nil {
			logrus.Errorf("failed to decode inAction(%+v): %v", inAction, err)
		} else {
			out.Actions = append(out.Actions, outAction)
		}
	}
	return out, nil
}

type TransactionAction struct {
	Header  *SignatureHeader        `json:"header,omitempty"`
	Payload *ChaincodeActionPayload `json:"payload,omitempty"`
}

func DecodeTransactionAction(in *peer.TransactionAction) (*TransactionAction, error) {
	if in == nil {
		return nil, nil
	}
	out := &TransactionAction{}
	outHeader, err := UnmarshalSignatureHeader(in.Header)
	if err != nil {
		logrus.Errorf("failed to unmarshal in.Header(%s): %v", in.Header, err)
	} else {
		out.Header = outHeader
	}
	outPayload, err := UnmarshalChaincodeActionPayload(in.Payload)
	if err != nil {
		logrus.Errorf("failed to unmarshal in.Payload(%s): %v", in.Payload, err)
	} else {
		out.Payload = outPayload
	}
	return out, nil
}

type ChaincodeActionPayload struct {
	ChaincodeProposalPayload *ChaincodeProposalPayload `json:"chaincode_proposal_payload,omitempty"`
	Action                   *ChaincodeEndorsedAction  `json:"action,omitempty"`
}

func UnmarshalChaincodeActionPayload(raw []byte) (*ChaincodeActionPayload, error) {
	in := &peer.ChaincodeActionPayload{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodeChaincodeActionPayload(in)
}

func DecodeChaincodeActionPayload(in *peer.ChaincodeActionPayload) (*ChaincodeActionPayload, error) {
	if in == nil {
		return nil, nil
	}
	out := &ChaincodeActionPayload{}
	outChaincodeProposalPayload, err := UnmarshalChaincodeProposalPayload(in.ChaincodeProposalPayload)
	if err != nil {
		logrus.Errorf("failed to unmarshal in.ChaincodeProposalPayload(%s): %v", in.ChaincodeProposalPayload, err)
	} else {
		out.ChaincodeProposalPayload = outChaincodeProposalPayload
	}
	outAction, err := DecodeChaincodeEndorsedAction(in.Action)
	if err != nil {
		logrus.Errorf("failed to decode in.Action(%+v): %v", in.Action, err)
	} else {
		out.Action = outAction
	}
	return out, nil
}

type ChaincodeEndorsedAction struct {
	ProposalResponsePayload *ProposalResponsePayload `json:"proposal_response_payload,omitempty"`
	Endorsements            []*Endorsement           `json:"endorsements,omitempty"`
}

func DecodeChaincodeEndorsedAction(in *peer.ChaincodeEndorsedAction) (*ChaincodeEndorsedAction, error) {
	if in == nil {
		return nil, nil
	}
	out := &ChaincodeEndorsedAction{}
	outProposalResponsePayload, err := UnmarshalProposalResponsePayload(in.ProposalResponsePayload)
	if err != nil {
		logrus.Errorf("failed to unmarshal in.ProposalResponsePayload(%s): %v", in.ProposalResponsePayload, err)
	} else {
		out.ProposalResponsePayload = outProposalResponsePayload
	}
	for _, inEndorsement := range in.Endorsements {
		outEndorsement, err := DecodeEndorsement(inEndorsement)
		if err != nil {
			logrus.Errorf("failed to decode inEndorsement(%+v): %v", inEndorsement, err)
		} else {
			out.Endorsements = append(out.Endorsements, outEndorsement)
		}
	}
	return out, nil
}

type Endorsement struct {
	Endorser  *SerializedIdentity `json:"endorser,omitempty"`
	Signature string              `json:"signature,omitempty"`
}

func DecodeEndorsement(in *peer.Endorsement) (*Endorsement, error) {
	if in == nil {
		return nil, nil
	}
	out := &Endorsement{
		Signature: hex.EncodeToString(in.Signature),
	}
	outEndorser, err := UnmarshalSerializedIdentity(in.Endorser)
	if err != nil {
		logrus.Errorf("failed unmarshal in.Endorser(%s): %v", in.Endorser, err)
	} else {
		out.Endorser = outEndorser
	}
	return out, nil
}

type ChaincodeProposalPayload struct {
	Input        *ChaincodeInvocationSpec `json:"input,omitempty"`
	TransientMap map[string][]byte        `json:"transient_map,omitempty"`
}

func UnmarshalChaincodeProposalPayload(raw []byte) (*ChaincodeProposalPayload, error) {
	in := &peer.ChaincodeProposalPayload{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodeChaincodeProposalPayload(in)
}

func DecodeChaincodeProposalPayload(in *peer.ChaincodeProposalPayload) (*ChaincodeProposalPayload, error) {
	if in == nil {
		return nil, nil
	}
	out := &ChaincodeProposalPayload{
		TransientMap: in.TransientMap,
	}
	outInput, err := UnmarshalChaincodeInvocationSpec(in.Input)
	if err != nil {
		logrus.Errorf("failed to unmarshal in.Input(%s): %v", in.Input, err)
	} else {
		out.Input = outInput
	}
	return out, nil
}

type ChaincodeInvocationSpec struct {
	ChaincodeSpec *ChaincodeSpec `json:"chaincode_spec,omitempty"`
}

func UnmarshalChaincodeInvocationSpec(raw []byte) (*ChaincodeInvocationSpec, error) {
	in := &peer.ChaincodeInvocationSpec{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodeChaincodeInvocationSpec(in)
}

func DecodeChaincodeInvocationSpec(in *peer.ChaincodeInvocationSpec) (*ChaincodeInvocationSpec, error) {
	if in == nil {
		return nil, nil
	}
	out := &ChaincodeInvocationSpec{}
	outChaincodeSpec, err := DecodeChaincodeSpec(in.ChaincodeSpec)
	if err != nil {
		logrus.Errorf("failed to decode in.ChaincodeSpec(%+v): %v", in.ChaincodeSpec, err)
	} else {
		out.ChaincodeSpec = outChaincodeSpec
	}
	return out, nil
}

type ChaincodeSpec struct {
	Type        string          `json:"type,omitempty"`
	ChaincodeId *ChaincodeID    `json:"chaincode_id,omitempty"`
	Input       *ChaincodeInput `json:"input,omitempty"`
	Timeout     int32           `json:"timeout,omitempty"`
}

func DecodeChaincodeSpec(in *peer.ChaincodeSpec) (*ChaincodeSpec, error) {
	if in == nil {
		return nil, nil
	}
	out := &ChaincodeSpec{
		Type:    in.Type.String(),
		Timeout: in.Timeout,
	}
	outChaincodeId, err := DecodeChaincodeID(in.ChaincodeId)
	if err != nil {
		logrus.Errorf("failed to decode in.ChaincodeId(%+v): %v", in.ChaincodeId, err)
	} else {
		out.ChaincodeId = outChaincodeId
	}
	outInput, err := DecodeChaincodeInput(in.Input)
	if err != nil {
		logrus.Errorf("failed to decode in.Input(%+v): %v", in.Input, err)
	} else {
		out.Input = outInput
	}
	return out, nil
}

type ChaincodeInput struct {
	Args        []string          `json:"args,omitempty"`
	Decorations map[string][]byte `json:"decorations,omitempty"`
	IsInit      bool              `json:"is_init,omitempty"`
}

func DecodeChaincodeInput(in *peer.ChaincodeInput) (*ChaincodeInput, error) {
	if in == nil {
		return nil, nil
	}
	out := &ChaincodeInput{
		Decorations: in.Decorations,
		IsInit:      in.IsInit,
	}
	for _, inArg := range in.Args {
		outArg := string(inArg)
		out.Args = append(out.Args, outArg)
	}
	return out, nil
}

type ProposalResponsePayload struct {
	ProposalHash    string           `json:"proposal_hash,omitempty"`
	ChaincodeAction *ChaincodeAction `json:"chaincode_action,omitempty"`
}

func UnmarshalProposalResponsePayload(raw []byte) (*ProposalResponsePayload, error) {
	in := &peer.ProposalResponsePayload{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodeProposalResponsePayload(in)
}

func DecodeProposalResponsePayload(in *peer.ProposalResponsePayload) (*ProposalResponsePayload, error) {
	if in == nil {
		return nil, nil
	}
	out := &ProposalResponsePayload{
		ProposalHash: hex.EncodeToString(in.ProposalHash),
	}
	outChaincodeAction, err := UnmarshalChaincodeAction(in.Extension)
	if err != nil {
		logrus.Errorf("failed to unmarshal in.Extension(%s): %v", in.Extension, err)
	} else {
		out.ChaincodeAction = outChaincodeAction
	}
	return out, nil
}

type ChaincodeAction struct {
	Results     *TxReadWriteSet `json:"results,omitempty"`
	Events      *ChaincodeEvent `json:"events,omitempty"`
	Response    *Response       `json:"response,omitempty"`
	ChaincodeId *ChaincodeID    `json:"chaincode_id,omitempty"`
}

func UnmarshalChaincodeAction(raw []byte) (*ChaincodeAction, error) {
	in := &peer.ChaincodeAction{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodeChaincodeAction(in)
}

func DecodeChaincodeAction(in *peer.ChaincodeAction) (*ChaincodeAction, error) {
	if in == nil {
		return nil, nil
	}
	out := &ChaincodeAction{}
	outResults, err := UnmarshalTxReadWriteSet(in.Results)
	if err != nil {
		logrus.Errorf("failed to unmarshal in.Results(%s): %v", in.Results, err)
	} else {
		out.Results = outResults
	}
	outEvents, err := UnmarshalChaincodeEvent(in.Events)
	if err != nil {
		logrus.Errorf("failed to unmarshal in.Events(%s): %v", in.Events, err)
	} else {
		out.Events = outEvents
	}
	outResponse, err := DecodeResponse(in.Response)
	if err != nil {
		logrus.Errorf("failed to decode in.Response(%+v): %v", in.Response, err)
	} else {
		out.Response = outResponse
	}
	outChaincodeId, err := DecodeChaincodeID(in.ChaincodeId)
	if err != nil {
		logrus.Errorf("failed to decode in.ChaincodeId(%+v): %v", in.ChaincodeId, err)
	} else {
		out.ChaincodeId = outChaincodeId
	}
	return out, nil
}

type Response struct {
	Status  int32  `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Payload []byte `json:"payload,omitempty"`
}

func DecodeResponse(in *peer.Response) (*Response, error) {
	if in == nil {
		return nil, nil
	}
	out := &Response{
		Status:  in.Status,
		Message: in.Message,
		Payload: in.Payload,
	}
	return out, nil
}

type ChaincodeID struct {
	Path    string `json:"path,omitempty"`
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

func DecodeChaincodeID(in *peer.ChaincodeID) (*ChaincodeID, error) {
	if in == nil {
		return nil, nil
	}
	out := &ChaincodeID{
		Path:    in.Path,
		Name:    in.Name,
		Version: in.Version,
	}
	return out, nil
}

type ChaincodeEvent struct {
	ChaincodeId string `json:"chaincode_id,omitempty"`
	TxId        string `json:"tx_id,omitempty"`
	EventName   string `json:"event_name,omitempty"`
	Payload     string `json:"payload,omitempty"`
}

func UnmarshalChaincodeEvent(raw []byte) (*ChaincodeEvent, error) {
	in := &peer.ChaincodeEvent{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodeChaincodeEvent(in)
}

func DecodeChaincodeEvent(in *peer.ChaincodeEvent) (*ChaincodeEvent, error) {
	if in == nil {
		return nil, nil
	}
	out := &ChaincodeEvent{
		ChaincodeId: in.ChaincodeId,
		TxId:        in.TxId,
		EventName:   in.EventName,
		Payload:     string(in.Payload),
	}
	return out, nil
}

type TxReadWriteSet struct {
	DataModel int32             `json:"data_model,omitempty"`
	NsRwset   []*NsReadWriteSet `json:"ns_rwset,omitempty"`
}

func UnmarshalTxReadWriteSet(raw []byte) (*TxReadWriteSet, error) {
	in := &rwset.TxReadWriteSet{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodeTxReadWriteSet(in)
}

func DecodeTxReadWriteSet(in *rwset.TxReadWriteSet) (*TxReadWriteSet, error) {
	if in == nil {
		return nil, nil
	}
	out := &TxReadWriteSet{
		DataModel: int32(in.DataModel),
	}
	for _, inNsRwset := range in.NsRwset {
		outNsRwset, err := DecodeNsReadWriteSet(inNsRwset)
		if err != nil {
			logrus.Errorf("failed to decode inNsRwset(%+v): %v", inNsRwset, err)
		} else {
			out.NsRwset = append(out.NsRwset, outNsRwset)
		}
	}
	return out, nil
}

type NsReadWriteSet struct {
	Namespace             string                          `json:"namespace,omitempty"`
	Rwset                 *KVRWSet                        `json:"rwset,omitempty"`
	CollectionHashedRwset []*CollectionHashedReadWriteSet `json:"collection_hashed_rwset,omitempty"`
}

func DecodeNsReadWriteSet(in *rwset.NsReadWriteSet) (*NsReadWriteSet, error) {
	if in == nil {
		return nil, nil
	}
	out := &NsReadWriteSet{
		Namespace: in.Namespace,
	}
	outRwset, err := UnmarshalKVRWSet(in.Rwset)
	if err != nil {
		logrus.Errorf("failed to unmarshal in.Rwset(%s): %v", in.Rwset, err)
	} else {
		out.Rwset = outRwset
	}
	for _, inCollectionHashedRwset := range in.CollectionHashedRwset {
		outCollectionHashedRwset, err := DecodeCollectionHashedReadWriteSet(inCollectionHashedRwset)
		if err != nil {
			logrus.Errorf("failed to decode inCollectionHashedRwset(%+v): %v", inCollectionHashedRwset, err)
		} else {
			out.CollectionHashedRwset = append(out.CollectionHashedRwset, outCollectionHashedRwset)
		}
	}
	return out, nil
}

type KVRWSet struct {
	Reads            []*KVRead          `json:"reads,omitempty"`
	RangeQueriesInfo []*RangeQueryInfo  `json:"range_queries_info,omitempty"`
	Writes           []*KVWrite         `json:"writes,omitempty"`
	MetadataWrites   []*KVMetadataWrite `json:"metadata_writes,omitempty"`
}

func UnmarshalKVRWSet(raw []byte) (*KVRWSet, error) {
	in := &kvrwset.KVRWSet{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodeKVRWSet(in)
}

func DecodeKVRWSet(in *kvrwset.KVRWSet) (*KVRWSet, error) {
	if in == nil {
		return nil, nil
	}
	out := &KVRWSet{}
	for _, inRead := range in.Reads {
		outRead, err := DecodeKVRead(inRead)
		if err != nil {
			logrus.Errorf("failed to decode inRead(%+v): %v", inRead, err)
		} else {
			out.Reads = append(out.Reads, outRead)
		}
	}
	for _, inWrite := range in.Writes {
		outWrite, err := DecodeKVWrite(inWrite)
		if err != nil {
			logrus.Errorf("failed to decode inWrite(%+v): %v", inWrite, err)
		} else {
			out.Writes = append(out.Writes, outWrite)
		}
	}
	for _, inRangeQueryInfo := range in.RangeQueriesInfo {
		outRangeQueryInfo, err := DecodeRangeQueryInfo(inRangeQueryInfo)
		if err != nil {
			logrus.Errorf("failed to decode inRangeQueryInfo(%+v): %v", inRangeQueryInfo, err)
		} else {
			out.RangeQueriesInfo = append(out.RangeQueriesInfo, outRangeQueryInfo)
		}
	}
	for _, inMetadataWrite := range in.MetadataWrites {
		outMetadataWrite, err := DecodeKVMetadataWrite(inMetadataWrite)
		if err != nil {
			logrus.Errorf("failed to decode inMetadataWrite(%+v): %v", inMetadataWrite, err)
		} else {
			out.MetadataWrites = append(out.MetadataWrites, outMetadataWrite)
		}
	}
	return out, nil
}

type KVRead struct {
	Key     string   `json:"key,omitempty"`
	Version *Version `json:"version,omitempty"`
}

func DecodeKVRead(in *kvrwset.KVRead) (*KVRead, error) {
	if in == nil {
		return nil, nil
	}
	out := &KVRead{
		Key: in.Key,
	}
	outVersion, err := DecodeVersion(in.Version)
	if err != nil {
		logrus.Errorf("failed to decode in.Version(%+v): %v", in.Version, err)
	} else {
		out.Version = outVersion
	}
	return out, nil
}

type Version struct {
	BlockNum uint64 `json:"block_num,omitempty"`
	TxNum    uint64 `json:"tx_num,omitempty"`
}

func DecodeVersion(in *kvrwset.Version) (*Version, error) {
	if in == nil {
		return nil, nil
	}
	out := &Version{
		BlockNum: in.BlockNum,
		TxNum:    in.TxNum,
	}
	return out, nil
}

type KVWrite struct {
	Key      string `json:"key,omitempty"`
	IsDelete bool   `json:"is_delete,omitempty"`
	Value    string `json:"value,omitempty"`
}

func DecodeKVWrite(in *kvrwset.KVWrite) (*KVWrite, error) {
	if in == nil {
		return nil, nil
	}
	out := &KVWrite{
		Key:      in.Key,
		IsDelete: in.IsDelete,
		Value:    string(in.Value),
	}
	return out, nil
}

type RangeQueryInfo struct {
	StartKey          string                   `json:"start_key,omitempty"`
	EndKey            string                   `json:"end_key,omitempty"`
	ItrExhausted      bool                     `json:"itr_exhausted,omitempty"`
	RawReads          *QueryReads              `json:"raw_reads,omitempty"`
	ReadsMerkleHashes *QueryReadsMerkleSummary `json:"reads_merkle_hashes,omitempty"`
}

func DecodeRangeQueryInfo(in *kvrwset.RangeQueryInfo) (*RangeQueryInfo, error) {
	if in == nil {
		return nil, nil
	}
	out := &RangeQueryInfo{
		StartKey:     in.EndKey,
		EndKey:       in.EndKey,
		ItrExhausted: in.ItrExhausted,
	}
	inRawReads := in.GetRawReads()
	outRawReads, err := DecodeQueryReads(inRawReads)
	if err != nil {
		logrus.Errorf("failed to decode inRawReads(%+v): %v", inRawReads, err)
	} else {
		out.RawReads = outRawReads
	}
	inReadsMerkleHashes := in.GetReadsMerkleHashes()
	outReadsMerkleHashes, err := DecodeQueryReadsMerkleSummary(inReadsMerkleHashes)
	if err != nil {
		logrus.Errorf("failed to decode inReadsMerkleHashes(%+v): %v", inReadsMerkleHashes, err)
	} else {
		out.ReadsMerkleHashes = outReadsMerkleHashes
	}
	return out, nil
}

type QueryReads struct {
	KvReads []*KVRead `json:"kv_reads,omitempty"`
}

func DecodeQueryReads(in *kvrwset.QueryReads) (*QueryReads, error) {
	if in == nil {
		return nil, nil
	}
	out := &QueryReads{}
	for _, inKvRead := range in.KvReads {
		outKvRead, err := DecodeKVRead(inKvRead)
		if err != nil {
			logrus.Errorf("failed to decode inKvRead(%+v): %v", inKvRead, err)
		} else {
			out.KvReads = append(out.KvReads, outKvRead)
		}
	}
	return out, nil
}

type QueryReadsMerkleSummary struct {
	MaxDegree      uint32   `json:"max_degree,omitempty"`
	MaxLevel       uint32   `json:"max_level,omitempty"`
	MaxLevelHashes [][]byte `json:"max_level_hashes,omitempty"`
}

func DecodeQueryReadsMerkleSummary(in *kvrwset.QueryReadsMerkleSummary) (*QueryReadsMerkleSummary, error) {
	if in == nil {
		return nil, nil
	}
	out := &QueryReadsMerkleSummary{
		MaxDegree:      in.MaxDegree,
		MaxLevel:       in.MaxLevel,
		MaxLevelHashes: in.MaxLevelHashes,
	}
	return out, nil
}

type KVMetadataWrite struct {
	Key     string             `json:"key,omitempty"`
	Entries []*KVMetadataEntry `json:"entries,omitempty"`
}

func DecodeKVMetadataWrite(in *kvrwset.KVMetadataWrite) (*KVMetadataWrite, error) {
	if in == nil {
		return nil, nil
	}
	out := &KVMetadataWrite{
		Key: in.Key,
	}
	for _, inEntry := range in.Entries {
		outEntry, err := DecodeKVMetadataEntry(inEntry)
		if err != nil {
			logrus.Errorf("failed to decode inEntry(%+v): %v", inEntry, err)
		} else {
			out.Entries = append(out.Entries, outEntry)
		}
	}
	return out, nil
}

type KVMetadataEntry struct {
	Name  string `json:"name,omitempty"`
	Value []byte `json:"value,omitempty"`
}

func DecodeKVMetadataEntry(in *kvrwset.KVMetadataEntry) (*KVMetadataEntry, error) {
	if in == nil {
		return nil, nil
	}
	out := &KVMetadataEntry{
		Name:  in.Name,
		Value: in.Value,
	}
	return out, nil
}

type CollectionHashedReadWriteSet struct {
	CollectionName string `json:"collection_name,omitempty"`
	HashedRwset    []byte `json:"hashed_rwset,omitempty"`
	PvtRwsetHash   []byte `json:"pvt_rwset_hash,omitempty"`
}

func DecodeCollectionHashedReadWriteSet(in *rwset.CollectionHashedReadWriteSet) (*CollectionHashedReadWriteSet, error) {
	if in == nil {
		return nil, nil
	}
	out := &CollectionHashedReadWriteSet{
		CollectionName: in.CollectionName,
		HashedRwset:    in.HashedRwset,
		PvtRwsetHash:   in.PvtRwsetHash,
	}
	return out, nil
}

type ChannelQueryResponse struct {
	Channels []*ChannelInfo `json:"channels,omitempty"`
}

func UnmarshalChannelQueryResponse(raw []byte) (*ChannelQueryResponse, error) {
	in := &peer.ChannelQueryResponse{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodeChannelQueryResponse(in)
}

func DecodeChannelQueryResponse(in *peer.ChannelQueryResponse) (*ChannelQueryResponse, error) {
	if in == nil {
		return nil, nil
	}
	out := &ChannelQueryResponse{}
	for _, inChannel := range in.Channels {
		outChannel, err := DecodeChannelInfo(inChannel)
		if err != nil {
			logrus.Errorf("failed to decode inChannel(%+v): %v", inChannel, err)
		} else {
			out.Channels = append(out.Channels, outChannel)
		}
	}
	return out, nil
}

type ChannelInfo struct {
	ChannelId string `json:"channel_id,omitempty"`
}

func UnmarshalChannelInfo(raw []byte) (*ChannelInfo, error) {
	in := &peer.ChannelInfo{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodeChannelInfo(in)
}

func DecodeChannelInfo(in *peer.ChannelInfo) (*ChannelInfo, error) {
	if in == nil {
		return nil, nil
	}
	out := &ChannelInfo{
		ChannelId: in.ChannelId,
	}
	return out, nil
}

type ChaincodeQueryResponse struct {
	Chaincodes []*ChaincodeInfo `json:"chaincodes,omitempty"`
}

func UnmarshalChaincodeQueryResponse(raw []byte) (*ChaincodeQueryResponse, error) {
	in := &peer.ChaincodeQueryResponse{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodeChaincodeQueryResponse(in)
}

func DecodeChaincodeQueryResponse(in *peer.ChaincodeQueryResponse) (*ChaincodeQueryResponse, error) {
	if in == nil {
		return nil, nil
	}
	out := &ChaincodeQueryResponse{}
	for _, inChaincode := range in.Chaincodes {
		outChaincode, err := DecodeChaincodeInfo(inChaincode)
		if err != nil {
			logrus.Errorf("failed to decode inChaincode(%+v): %v", inChaincode, err)
		} else {
			out.Chaincodes = append(out.Chaincodes, outChaincode)
		}
	}
	return out, nil
}

type ChaincodeInfo struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
	Path    string `json:"path,omitempty"`
	Input   string `json:"input,omitempty"`
	Escc    string `json:"escc,omitempty"`
	Vscc    string `json:"vscc,omitempty"`
	Id      string `json:"id,omitempty"`
}

func UnmarshalChaincodeInfo(raw []byte) (*ChaincodeInfo, error) {
	in := &peer.ChaincodeInfo{}
	err := proto.Unmarshal(raw, in)
	if err != nil {
		return nil, err
	}
	return DecodeChaincodeInfo(in)
}

func DecodeChaincodeInfo(in *peer.ChaincodeInfo) (*ChaincodeInfo, error) {
	if in == nil {
		return nil, nil
	}
	out := &ChaincodeInfo{
		Name:    in.Name,
		Version: in.Version,
		Path:    in.Path,
		Input:   in.Input,
		Escc:    in.Escc,
		Vscc:    in.Vscc,
		Id:      hex.EncodeToString(in.Id),
	}
	return out, nil
}

// InstallCCResponse contains install chaincode response status
type InstallCCResponse struct {
	Target string `json:"target,omitempty"`
	Status int32  `json:"status,omitempty"`
	Info   string `json:"info,omitempty"`
}

func DecodeInstallCCResponse(in []resmgmt.InstallCCResponse) ([]InstallCCResponse, error) {
	var out []InstallCCResponse
	if len(in) == 0 {
		return nil, nil
	}
	for i := range in {
		out = append(out, InstallCCResponse{
			Target: in[i].Target,
			Status: in[i].Status,
			Info:   in[i].Info,
		})
	}
	return out, nil
}

// InstantiateCCResponse contains response parameters for instantiate chaincode
type InstantiateCCResponse struct {
	TransactionID string `json:"transactionID,omitempty"`
}

func DecodeInstantiateCCResponse(in resmgmt.InstantiateCCResponse) (*InstantiateCCResponse, error) {
	out := &InstantiateCCResponse{
		TransactionID: string(in.TransactionID),
	}
	return out, nil
}

// SaveChannelResponse contains response parameters for save channel
type SaveChannelResponse struct {
	TransactionID string `json:"transactionID,omitempty"`
}

func DecodeSaveChannelResponse(in resmgmt.SaveChannelResponse) (*SaveChannelResponse, error) {
	out := &SaveChannelResponse{
		TransactionID: string(in.TransactionID),
	}
	return out, nil
}

// UpgradeCCResponse contains response parameters for upgrade chaincode
type UpgradeCCResponse struct {
	TransactionID string `json:"transactionID,omitempty"`
}

func DecodeUpgradeCCResponse(in resmgmt.UpgradeCCResponse) (*UpgradeCCResponse, error) {
	out := &UpgradeCCResponse{
		TransactionID: string(in.TransactionID),
	}
	return out, nil
}

// InvokeChainCode
type InvokeChainCodeResponse struct {
	Proposal         Proposal           `json:"Proposal"`
	Responses        []EndorserResponse `json:"Responses"`
	TransactionID    string             `json:"TransactionID"`
	TxValidationCode int64              `json:"TxValidationCode"`
	ChaincodeStatus  int64              `json:"ChaincodeStatus"`
	Payload          string             `json:"Payload"`
	PayloadData      []byte             `json:"-"`
}

type Proposal struct {
	TxnID   string `json:"TxnID"`
	Header  string `json:"header"`
	Payload string `json:"payload"`
}

type EndorserResponse struct {
	Endorser        string            `json:"Endorser"`
	Status          int64             `json:"Status"`
	ChainCodeStatus int64             `json:"ChaincodeStatus"`
	Version         int32             `json:"version"`
	Response        ChainCodeResponse `json:"response"`
	Payload         string            `json:"payload"`
	Endorsement     EndorsementEncode `json:"endorsement"`
}

type ChainCodeResponse struct {
	Status  int64  `json:"status"`
	Payload string `json:"payload"`
}

type EndorsementEncode struct {
	Endorser  string `json:"endorser"`
	Signature string `json:"signature"`
}

func DecodeInvokeChainCodeResponse(payload []byte) (*InvokeChainCodeResponse, error) {
	var invokeChainCodeResponse InvokeChainCodeResponse
	err := json.Unmarshal(payload, &invokeChainCodeResponse)
	if err != nil {
		return nil, err
	}

	// 这里暂时只对payload的内容进行了decode
	data, err := base64.StdEncoding.DecodeString(invokeChainCodeResponse.Payload)
	if err != nil {
		return nil, err
	}
	invokeChainCodeResponse.PayloadData = data
	return &invokeChainCodeResponse, nil
}
