// Copyright © 2025 Ping Identity Corporation

// Package main provides a CLI plugin for converting PingOne DaVinci flows
// (in JSON format) to HCL (HashiCorp Configuration Language) that is compatible
// with the PingOne Terraform Provider's DaVinci resources.
package main

import (
	"github.com/hashicorp/go-plugin"
	"github.com/pingidentity/pingcli/shared/grpc"
	"github.com/samir-gandhi/davinci-terraform-converter/cmd"
)

// main is the entry point for the plugin's executable. When the pingcli host
// launches this plugin, this function starts a gRPC server that serves the
// DaVinciConvertCommand implementation.
func main() {
	plugin.Serve(&plugin.ServeConfig{
		// HandshakeConfig is a shared configuration used to verify that the host
		// and plugin are compatible.
		HandshakeConfig: grpc.HandshakeConfig,

		// Plugins defines the set of services this plugin serves. The key is a
		// unique name for the plugin service, and the value is an implementation
		// of the plugin.Plugin interface.
		Plugins: map[string]plugin.Plugin{
			grpc.ENUM_PINGCLI_COMMAND_GRPC: &grpc.PingCliCommandGrpcPlugin{
				Impl: &cmd.DaVinciConvertCommand{},
			},
		},

		// GRPCServer specifies the gRPC server implementation to use.
		// plugin.DefaultGRPCServer is a sane default provided by the library.
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
