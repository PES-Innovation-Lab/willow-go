package transport

func client(){
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 3s handshake timeout
	defer cancel()
	conn, err := tr.Dial(ctx, <server address>, <tls.Config>, <quic.Config>)
	// ... error handling
}

