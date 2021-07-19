package template

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/rmohr/bazeldnf/pkg/api"
)

func Render(writer io.Writer, installed []*api.Package, forceIgnored []*api.Package) error {
	totalDownloadSize := 0
	totalInstallSize := 0

	tabWriter := tabwriter.NewWriter(writer, 0, 8, 1, '\t', 0)
	if _, err := fmt.Fprintln(tabWriter, "Package\tVersion\tSize\tDownload Size"); err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}
	if _, err := fmt.Fprintln(tabWriter, "Installing:\t\t\t"); err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}
	for _, pkg := range installed {
		totalInstallSize += pkg.Size.Archive
		totalDownloadSize += pkg.Size.Package
		if _, err := fmt.Fprintf(tabWriter, " %v\t%v\t%s\t%s\n", pkg.Name, pkg.Version.String(), toReadableQuantity(pkg.Size.Archive), toReadableQuantity(pkg.Size.Package)); err != nil {
			return fmt.Errorf("failed to write entry: %v", err)
		}
	}
	if _, err := fmt.Fprintln(tabWriter, "Ignoring:\t\t\t"); err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}
	for _, pkg := range forceIgnored {
		if _, err := fmt.Fprintf(tabWriter, " %v\t%v\t%s\t%s\n", pkg.Name, pkg.Version.String(), toReadableQuantity(pkg.Size.Archive), toReadableQuantity(pkg.Size.Package)); err != nil {
			return fmt.Errorf("failed to write entry: %v", err)
		}
	}
	if _, err := fmt.Fprintln(tabWriter, "\t\t\t\nTransaction Summary:\t\t\t"); err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}
	if _, err := fmt.Fprintf(tabWriter, "Installing %d Packages \t\t\t\n", len(installed)); err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}
	if _, err := fmt.Fprintf(tabWriter, "Total download size: %s\t\t\t\n", toReadableQuantity(totalDownloadSize)); err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}
	if _, err := fmt.Fprintf(tabWriter, "Total install size: %s\t\t\t\n", toReadableQuantity(totalInstallSize)); err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}
	if err := tabWriter.Flush(); err != nil {
		return fmt.Errorf("failed to flush table: %v", err)
	}
	return nil
}

func toReadableQuantity(bytes int) string {
	if bytes > 1000*1000*1000 {
		q := float64(bytes) / 1000 / 1000 / 1000
		return fmt.Sprintf("%.2f G", q)
	} else if bytes > 1000*1000 {
		q := float64(bytes) / 1000 / 1000
		return fmt.Sprintf("%.2f M", q)
	} else if bytes > 1000 {
		q := float64(bytes) / 1000
		return fmt.Sprintf("%.2f K", q)
	} else {
		return fmt.Sprintf("%d", bytes)
	}
}
