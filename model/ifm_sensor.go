package model

const IFMDataKey = 0x10020001

type IFMData struct {
	PortTimestamp         int64  `json:"port_timestamp"`
	Ifx                   int    `json:"ifx"`
	PortName              string `json:"port_name"`
	InpErrorPkts          int64  `json:"inp_error_pkts"`
	OutpErrorPkts         int64  `json:"outp_error_pkts"`
	InpDropPkts           int64  `json:"inp_drop_pkts"`
	OutpDropPkts          int64  `json:"outp_drop_pkts"`
	InpUcastPkts          int64  `json:"inp_ucast_pkts"`
	OutpUcastPkts         int64  `json:"outp_ucast_pkts"`
	IfInOctets            int64  `json:"if_in_octets"`
	IfOutOctets           int64  `json:"if_out_octets"`
	TotalDiscardPkts      int64  `json:"total_discard_pkts"`
	RxAverRate            int    `json:"rx_aver_rate"`
	RxAverPktRate         int    `json:"rx_aver_pkt_rate"`
	TxAverRate            int    `json:"tx_aver_rate"`
	TxAverPktRate         int    `json:"tx_aver_pkt_rate"`
	IfInOctetsKb          int64  `json:"if_in_octets_kb"`
	IfOutOctetsKb         int64  `json:"if_out_octets_kb"`
	InpPkts               int64  `json:"inp_pkts"`
	OutpPkts              int64  `json:"outp_pkts"`
	OutpMultiPkts         int64  `json:"outp_multi_pkts"`
	OutpBroadPkts         int64  `json:"outp_broad_pkts"`
	InpMultiPkts          int64  `json:"inp_multi_pkts"`
	InpBroadPkts          int64  `json:"inp_broad_pkts"`
	InpCrcerrorPkts       int64  `json:"inp_crcerror_pkts"`
	InpNUcastPkts         int64  `json:"inp_nucast_pkts"`
	OutpNUcastPkts        int64  `json:"outp_nucast_pkts"`
	InpNobufferPkts       int64  `json:"inp_nobuffer_pkts"`
	OutpNobufferPkts      int64  `json:"outp_nobuffer_pkts"`
	InpDiscardPkts        int64  `json:"inp_discard_pkts"`
	OutpDiscardPkts       int64  `json:"outp_discard_pkts"`
	InpPausePkts          int64  `json:"inp_pause_pkts"`
	OutpPausePkts         int64  `json:"outp_pause_pkts"`
	InpOversizePkts       int64  `json:"inp_oversize_pkts"`
	InpJabberPkts         int64  `json:"inp_jabber_pkts"`
	InpFragmentPkts       int64  `json:"inp_fragment_pkts"`
	InpUndersizePkts      int64  `json:"inp_undersize_pkts"`
	InpJumboPkts          int64  `json:"inp_jumbo_pkts"`
	OutpJumboPkts         int64  `json:"outp_jumbo_pkts"`
	InpNobufferPktsDelta  int64  `json:"inp_nobuffer_pkts_delta"`
	OutpNobufferPktsDelta int64  `json:"outp_nobuffer_pkts_delta"`
}
