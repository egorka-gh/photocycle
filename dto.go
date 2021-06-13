package photocycle

//PrintPostedEFI repo DTO
type PrintPostedEFI struct {
	PrintgroupID string `db:"id"`
	FilesCount   int    `db:"fileCount"`
}
