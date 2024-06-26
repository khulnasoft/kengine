package controls

import (
	ctl "github.com/khulnasoft/kengine/khulnasoft_utils/controls"
	"github.com/khulnasoft/kengine/khulnasoft_utils/log"
	k8sscanner "github.com/khulnasoft-lab/kubernetes-scanner/v2/scanner/compliance"
	k8sscannerutil "github.com/khulnasoft-lab/kubernetes-scanner/v2/util"
)

func StartComplianceScan(req ctl.StartComplianceScanRequest) error {
	scanner, err := k8sscanner.NewComplianceScanner(
		k8sscannerutil.Config{
			ComplianceCheckType:       k8sscannerutil.NsaCisaCheckType,
			ScanId:                    req.BinArgs["scan_id"],
			NodeId:                    req.NodeID,
			NodeName:                  req.NodeID,
			ComplianceResultsFilePath: "/var/log/fenced/compliance/compliance-scan.log",
			ComplianceStatusFilePath:  "/var/log/fenced/compliance-scan-logs/status.log",
		})
	if err != nil {
		return err
	}
	err = scanner.RunComplianceScan()
	if err != nil {
		log.Error().Msgf("Error from scan: %+v", err)
		return err
	}
	return nil
}

func StartClusterAgentUpgrade(req ctl.StartAgentUpgradeRequest) error {
	return nil
}
