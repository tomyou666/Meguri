package memprobe

// AvailableBytes は OS 上で利用可能な物理メモリ量（バイト）を返す。
func AvailableBytes() (uint64, error) {
	return platformAvailableBytes()
}

// UsedRatio は使用中メモリ / 総メモリの比率を返す。
func UsedRatio() (float64, error) {
	return platformUsedRatio()
}
