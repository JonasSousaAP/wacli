package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/openclaw/wacli/internal/out"
	"github.com/spf13/cobra"
	"go.mau.fi/whatsmeow/types"
)

func newGroupsSetPictureCmd(flags *rootFlags) *cobra.Command {
	var jidStr string
	cmd := &cobra.Command{
		Use:   "set-picture <image>",
		Short: "Set a group's picture (JPEG or PNG, auto-resized to <=640px)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := flags.requireWritable(); err != nil {
				return err
			}
			jid, err := types.ParseJID(strings.TrimSpace(jidStr))
			if err != nil || jid.Server != types.GroupServer {
				return fmt.Errorf("--jid must be a group JID (…@g.us)")
			}
			imgBytes, err := readAsJPEG(args[0])
			if err != nil {
				return fmt.Errorf("read image: %w", err)
			}
			ctx, cancel := withTimeout(context.Background(), flags)
			defer cancel()
			a, lk, err := newApp(ctx, flags, true, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)
			if err := a.EnsureAuthed(ctx); err != nil {
				return err
			}
			if err := a.Connect(ctx, false, nil); err != nil {
				return err
			}
			id, err := a.WA().SetGroupPicture(ctx, jid, imgBytes)
			if err != nil {
				return fmt.Errorf("set group picture: %w", err)
			}
			if flags.asJSON {
				return out.WriteJSON(os.Stdout, map[string]any{"group": jid.String(), "picture_id": id})
			}
			fmt.Fprintf(os.Stdout, "Group picture updated (id: %s)\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&jidStr, "jid", "", "group JID (…@g.us)")
	_ = cmd.MarkFlagRequired("jid")
	return cmd
}
