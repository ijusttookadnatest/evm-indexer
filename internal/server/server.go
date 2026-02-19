package server


type Server struct {
	server       *http.Server
}

func NewServer(port int, service ports.QueryService) *Server {
	return &Server{
		server: &http.Server{
			ReadTimeout: 10 * time.Second,
			WriteTimeout: 10 * time.Second,
			Addr:fmt.Sprintf(":%v", port),
			Handler: newRouter(service),
		},
	}
}

func (server *Server) Run() error {
	if err := server.server.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

// mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
// 	w.WriteHeader(http.StatusOK)
// 	fmt.Fprintf(w, "ok")
// })