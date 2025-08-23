package cmd

import (
	"github.com/spf13/cobra"
)

var workoutCmd = &cobra.Command{
	Use:   "workout",
	Short: "Track and view workouts",
	Long:  "Track and view workouts including viewing next workout and logging completed workouts.",
}

func init() {
	rootCmd.AddCommand(workoutCmd)
	workoutCmd.AddCommand(workoutNextCmd)
	workoutCmd.AddCommand(workoutLogCmd)
}

