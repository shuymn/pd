package cli

// Root is the kong root struct for the pd CLI.
type Root struct {
	Root string  `help:"Directory to scan, relative to repository root." default:"docs" name:"root"`
	List ListCmd `cmd:"" help:"List discovery metadata from docs directory."`
}
