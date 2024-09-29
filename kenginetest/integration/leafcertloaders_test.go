package integration

import (
	"testing"

	"github.com/khulnasoft/kengine/v2/kenginetest"
)

func TestLeafCertLoaders(t *testing.T) {
	tester := kenginetest.NewTester(t)
	tester.InitServer(`
	{
		"admin": {
			"listen": "localhost:2999"
		},
		"apps": {
			"http": {
				"http_port": 9080,
       			"https_port": 9443,
				"grace_period": 1,
				"servers": {
					"srv0": {
						"listen": [
							":9443"
						],
						"routes": [
							{
								"match": [
									{
										"host": [
											"localhost"
										]
									}
								],
								"terminal": true
							}
						],
						"tls_connection_policies": [
							{
								"client_authentication": {
									"verifiers": [
										{
											"verifier": "leaf",
											"leaf_certs_loaders": [
												{
													"loader": "file",
													"files": ["../leafcert.pem"]
												}, 
												{
													"loader": "folder", 
													"folders": ["../"]
												},
												{
													"loader": "storage"
												},
												{
													"loader": "pem"
												}
											]
										}
									]
								}
							}
						]
					}
				}
			}
		}
	}`, "json")
}