package handler

// 2021-10-28 增加心跳检测阈值判定
// func ParseHeartBeat(hb map[string]string, req *pb.RawData) {

// 	agentID := req.AgentID
// 	conn, err := GlobalGRPCPool.Get(agentID)
// 	if err != nil {
// 		return
// 	}

// 	clearConn(conn)

// 	strCPU, ok := hb["cpu"]
// 	if ok {
// 		if cpu, err := strconv.ParseFloat(strCPU, 64); err == nil {
// 			conn.Cpu = cpu
// 		}
// 	}

// 	strIO, ok := hb["io"]
// 	if ok {
// 		if io, err := strconv.ParseFloat(strIO, 64); err == nil {
// 			conn.IO = io
// 		}
// 	}

// 	strMem, ok := hb["memory"]
// 	if ok {
// 		if mem, err := strconv.ParseInt(strMem, 10, 64); err == nil {
// 			conn.Memory = mem
// 		}
// 	}

// 	strSlab, ok := hb["slab"]
// 	if ok {
// 		if slab, err := strconv.ParseInt(strSlab, 10, 64); err == nil {
// 			conn.Slab = slab
// 		}
// 	}

// 	conn.HostName = req.Hostname
// 	conn.Version = req.Version
// 	if req.IntranetIPv4 != nil {
// 		conn.IntranetIPv4 = req.IntranetIPv4
// 	}

// 	if req.IntranetIPv6 != nil {
// 		conn.IntranetIPv6 = req.IntranetIPv6
// 	}

// 	//last heartbeat time get from server
// 	conn.LastHeartBeatTime = time.Now().Unix()

// 	// 传输时间检测
// 	fmt.Printf("%f, %d\n", conn.Cpu, conn.Memory)
// }
