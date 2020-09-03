package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"google.golang.org/grpc"
	"v2ray.com/core"
	"v2ray.com/core/app/proxyman/command"
	"v2ray.com/core/infra/conf"
)

// APIServer represent a V2Ray API server
type APIServer struct {
	Host string
	Port uint16
	conn *grpc.ClientConn
}

var errNilResponse = errors.New("unexpected nil response")

// Conn returns an active connect to server
func (s *APIServer) getConn() (*grpc.ClientConn, error) {
	if s.conn == nil {
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", s.Host, s.Port), grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(10*time.Second))
		if err != nil {
			return nil, err
		}
		s.conn = conn
	}
	return s.conn, nil
}

// Close closes connect to server
func (s *APIServer) Close() {
	if s.conn == nil {
		return
	}
	s.conn.Close()
	s.conn = nil
}

// RemoveOutbounds remove outbounds by tags
func (s *APIServer) RemoveOutbounds(tags []string) error {
	if len(tags) == 0 {
		return nil
	}
	conn, err := s.getConn()
	if err != nil {
		return err
	}
	hsClient := command.NewHandlerServiceClient(conn)
	for _, tag := range tags {
		resp, err := hsClient.RemoveOutbound(context.Background(), &command.RemoveOutboundRequest{
			Tag: tag,
		})
		if err != nil {
			return err
		}
		if resp == nil {
			return errNilResponse
		}
	}
	return nil
}

// RemoveOutboundFiles remove outbounds by tags from files
func (s *APIServer) RemoveOutboundFiles(files []string) error {
	tags, err := filesToTags(files)
	if err != nil {
		return err
	}
	return s.RemoveOutbounds(tags)
}

// AddOutbounds add outbounds to server
func (s *APIServer) AddOutbounds(outbounds []*core.OutboundHandlerConfig) error {
	conn, err := s.getConn()
	if err != nil {
		return err
	}
	hsClient := command.NewHandlerServiceClient(conn)
	for _, outbound := range outbounds {
		resp, err := hsClient.AddOutbound(context.Background(), &command.AddOutboundRequest{
			Outbound: outbound,
		})
		if err != nil {
			return err
		}
		if resp == nil {
			return errNilResponse
		}
	}
	return nil
}

// AddOutboundFiles add outbounds from config files to server
func (s *APIServer) AddOutboundFiles(files []string) error {
	outbounds := make([]*core.OutboundHandlerConfig, 0)
	for _, file := range files {
		outs, err := jsonToOutboundHandlerConfigs(file)
		if err != nil {
			return err
		}
		outbounds = append(outbounds, outs...)
	}
	return s.AddOutbounds(outbounds)
}
func jsonToOutboundHandlerConfigs(f string) ([]*core.OutboundHandlerConfig, error) {
	confs, err := jsonToOutboundConfigs(f)
	if err != nil {
		return nil, err
	}
	if confs == nil || len(confs) == 0 {
		return nil, fmt.Errorf("no valid outbound found in %s", f)
	}
	outbounds := make([]*core.OutboundHandlerConfig, 0)
	for _, c := range confs {
		out, err := c.Build()
		if err != nil {
			return nil, err
		}
		outbounds = append(outbounds, out)
	}
	return outbounds, nil
}

func jsonToOutboundConfigs(f string) ([]conf.OutboundDetourConfig, error) {
	c := &conf.Config{}
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(data, c)
	if err != nil {
		return nil, err
	}
	return c.OutboundConfigs, nil
}

func filesToTags(files []string) ([]string, error) {
	tags := make([]string, 0)
	for _, file := range files {
		ts, err := fileToTags(file)
		if err != nil {
			return nil, err
		}
		tags = append(tags, ts...)
	}
	return tags, nil
}

func fileToTags(file string) ([]string, error) {
	tags := make([]string, 0)
	confs, err := jsonToOutboundConfigs(file)
	if err != nil {
		return nil, err
	}
	for _, c := range confs {
		tags = append(tags, c.Tag)
	}
	return tags, nil
}
