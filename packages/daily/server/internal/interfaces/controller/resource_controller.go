package controller

import (
	"context"
	"daily/internal/application/dto"
	"daily/internal/application/usecase/resource"
	"io"
)

type ResourceController struct {
	uploadUC *resource.UploadResourceUseCase
	getUC    *resource.GetResourceUseCase
	exportUC *resource.ExportDataUseCase
	importUC *resource.ImportDataUseCase
}

func NewResourceController(
	uploadUC *resource.UploadResourceUseCase,
	getUC *resource.GetResourceUseCase,
	exportUC *resource.ExportDataUseCase,
	importUC *resource.ImportDataUseCase,
) *ResourceController {
	return &ResourceController{
		uploadUC: uploadUC,
		getUC:    getUC,
		exportUC: exportUC,
		importUC: importUC,
	}
}

func (ctrl *ResourceController) Import(ctx context.Context, userID int64, r io.ReaderAt, size int64) (*dto.ImportReport, error) {
	return ctrl.importUC.Execute(ctx, userID, r, size)
}

func (ctrl *ResourceController) Upload(ctx context.Context, userID int64, filename, mimeType string, size int64, content io.Reader) (*dto.ResourceResponse, error) {
	input := dto.UploadResourceInput{
		FileName: filename,
		MimeType: mimeType,
		Size:     size,
		Content:  content,
	}
	return ctrl.uploadUC.Execute(ctx, userID, input)
}

func (ctrl *ResourceController) Get(ctx context.Context, userID int64, id string) (*resource.GetResourceOutput, error) {
	return ctrl.getUC.Execute(ctx, userID, id)
}

func (ctrl *ResourceController) Export(ctx context.Context, userID int64, w io.Writer) error {
	return ctrl.exportUC.Execute(ctx, userID, w)
}
