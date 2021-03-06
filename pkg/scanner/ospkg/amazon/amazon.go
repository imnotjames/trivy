package amazon

import (
	"strings"

	"go.uber.org/zap"

	"github.com/aquasecurity/fanal/analyzer"
	"github.com/aquasecurity/trivy/pkg/vulnsrc/amazon"
	version "github.com/knqyf263/go-deb-version"
	"golang.org/x/xerrors"

	"github.com/aquasecurity/trivy/pkg/log"
	"github.com/aquasecurity/trivy/pkg/scanner/utils"
	"github.com/aquasecurity/trivy/pkg/vulnsrc/vulnerability"
)

type Scanner struct {
	l  *zap.SugaredLogger
	ac amazon.Operations
}

func NewScanner() *Scanner {
	return &Scanner{
		l:  log.Logger,
		ac: amazon.NewVulnSrc(),
	}
}

func (s *Scanner) Detect(osVer string, pkgs []analyzer.Package) ([]vulnerability.DetectedVulnerability, error) {
	log.Logger.Info("Detecting Amazon Linux vulnerabilities...")

	osVer = strings.Fields(osVer)[0]
	if osVer != "2" {
		osVer = "1"
	}
	log.Logger.Debugf("amazon: os version: %s", osVer)
	log.Logger.Debugf("amazon: the number of packages: %d", len(pkgs))

	var vulns []vulnerability.DetectedVulnerability
	for _, pkg := range pkgs {
		advisories, err := s.ac.Get(osVer, pkg.Name)
		if err != nil {
			return nil, xerrors.Errorf("failed to get amazon advisories: %w", err)
		}

		installed := utils.FormatVersion(pkg)
		if installed == "" {
			continue
		}

		installedVersion, err := version.NewVersion(installed)
		if err != nil {
			log.Logger.Debugf("failed to parse Amazon Linux installed package version: %s", err)
			continue
		}

		for _, adv := range advisories {
			fixedVersion, err := version.NewVersion(adv.FixedVersion)
			if err != nil {
				log.Logger.Debugf("failed to parse Amazon Linux package version: %s", err)
				continue
			}

			if installedVersion.LessThan(fixedVersion) {
				vuln := vulnerability.DetectedVulnerability{
					VulnerabilityID:  adv.VulnerabilityID,
					PkgName:          pkg.Name,
					InstalledVersion: installed,
					FixedVersion:     adv.FixedVersion,
				}
				vulns = append(vulns, vuln)
			}
		}
	}
	return vulns, nil
}

func (s *Scanner) IsSupportedVersion(osFamily, osVer string) bool {
	return true
}
