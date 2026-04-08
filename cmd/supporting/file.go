package supporting

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/client"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
	"github.com/uqpay/uqpay-cli/internal/output"
)

func NewFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "Manage files (upload and download links)",
	}
	cmd.AddCommand(
		newFileUploadCmd(),
		newFileDownloadLinksCmd(),
	)
	return cmd
}

func newFileUploadCmd() *cobra.Command {
	var notes, onBehalfOf string
	cmd := &cobra.Command{
		Use:   "upload <filepath>",
		Short: "Upload a file to UQPAY",
		Long: `Upload a file to UQPAY. Returns a file_id that can be used in other endpoints.

Supported file types: jpeg, png, jpg, doc, docx, pdf (max 20 MB).

The file_id returned can be referenced in:
  - account create-sub (identity_verification, proof_documents fields)
  - other endpoints that accept document attachments

Examples:
  uqpay file upload ./passport.png
  uqpay file upload ./document.pdf --notes "passport front"
  uqpay file upload ./id.jpg --on-behalf-of <sub-account-id>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.NewFileClient(cfg)
			query := map[string]string{"notes": notes}
			headers := map[string]string{"x-on-behalf-of": onBehalfOf}
			data, err := c.PostMultipartH(context.Background(), "/v1/files/upload", args[0], query, headers)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&notes, "notes", "", "Notes for the uploaded file (max 50 chars)")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}

func newFileDownloadLinksCmd() *cobra.Command {
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "download-links <file-id> [<file-id>...]",
		Short: "Get download links for uploaded files",
		Long: `Get download links for one or more files by their file IDs.

The file IDs are returned by "file upload".

Examples:
  uqpay file download-links b3d9d2d5-4c12-4946-a09d-953e82sed2b0
  uqpay file download-links <id1> <id2> <id3>`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			fileIDs := make([]any, len(args))
			for i, id := range args {
				fileIDs[i] = id
			}
			body := map[string]any{"file_ids": fileIDs}
			c := client.NewFileClient(cfg)
			data, err := c.PostH(context.Background(), "/v1/files/download_links", body,
				map[string]string{"x-on-behalf-of": onBehalfOf})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}
