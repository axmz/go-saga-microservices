package http

import (
	"log/slog"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func RespondJSON(w http.ResponseWriter, data proto.Message, statusCode int) {
	jsonData, err := protojson.Marshal(data)
	if err != nil {
		slog.Error("Marshal JSON response failed", "err", err)
		ErrorInternal(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write(jsonData); err != nil {
		slog.Warn("Write JSON response failed", "err", err)
	}
}

func RespondProto(w http.ResponseWriter, data proto.Message, statusCode int) {
	bytes, err := proto.Marshal(data)
	if err != nil {
		slog.Error("Marshal proto response failed", "err", err)
		ErrorInternal(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/x-protobuf")
	w.WriteHeader(statusCode)
	if _, err := w.Write(bytes); err != nil {
		slog.Warn("Write proto response failed", "err", err)
	}
}
