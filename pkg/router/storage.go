package router

import (
	"net/http"

	"git.containerum.net/ch/volume-manager/pkg/errors"
	"git.containerum.net/ch/volume-manager/pkg/models"
	"git.containerum.net/ch/volume-manager/pkg/server"
	kubeClientModel "github.com/containerum/kube-client/pkg/model"
	"github.com/containerum/utils/httputil"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/sirupsen/logrus"
)

type storageHandlers struct {
	tv   *TranslateValidate
	acts server.StorageActions
}

func (sh *storageHandlers) createStorageHandler(ctx *gin.Context) {
	var req model.Storage
	if err := ctx.ShouldBindWith(&req, binding.JSON); err != nil {
		ctx.AbortWithStatusJSON(sh.tv.BadRequest(ctx, err))
		return
	}
	if err := sh.acts.CreateStorage(ctx.Request.Context(), req); err != nil {
		ctx.AbortWithStatusJSON(sh.tv.HandleError(err))
		return
	}

	ctx.Status(http.StatusCreated)
}

func (sh *storageHandlers) importStoragesHandler(ctx *gin.Context) {
	var req []string
	if err := ctx.ShouldBindWith(&req, binding.JSON); err != nil {
		ctx.AbortWithStatusJSON(sh.tv.BadRequest(ctx, err))
		return
	}

	resp := kubeClientModel.ImportResponse{
		Imported: []kubeClientModel.ImportResult{},
		Failed:   []kubeClientModel.ImportResult{},
	}

	for _, r := range req {
		store := model.Storage{
			Name: r,
			Size: 100,
		}

		if err := sh.acts.CreateStorage(ctx.Request.Context(), store); err != nil {
			logrus.Warn(err)
			resp.ImportFailed(r, "", err.Error())
		} else {
			resp.ImportSuccessful(r, "")
		}
	}

	ctx.JSON(http.StatusAccepted, resp)
}

func (sh *storageHandlers) getStoragesHandler(ctx *gin.Context) {
	storages, err := sh.acts.GetStorages(ctx.Request.Context())
	if err != nil {
		ctx.AbortWithStatusJSON(sh.tv.HandleError(err))
		return
	}

	ctx.JSON(http.StatusOK, storages)
}

func (sh *storageHandlers) updateStorageHandler(ctx *gin.Context) {
	var req model.UpdateStorageRequest
	if err := ctx.ShouldBindWith(&req, binding.JSON); err != nil {
		ctx.AbortWithStatusJSON(sh.tv.BadRequest(ctx, err))
		return
	}
	if err := sh.acts.UpdateStorage(ctx.Request.Context(), ctx.Param("name"), req); err != nil {
		ctx.AbortWithStatusJSON(sh.tv.HandleError(err))
		return
	}
	ctx.Status(http.StatusAccepted)
}

func (sh *storageHandlers) deleteStorageHandler(ctx *gin.Context) {
	if err := sh.acts.DeleteStorage(ctx.Request.Context(), ctx.Param("name")); err != nil {
		ctx.AbortWithStatusJSON(sh.tv.HandleError(err))
		return
	}
	ctx.Status(http.StatusAccepted)
}

func (r *Router) SetupStorageHandlers(acts server.StorageActions) {
	handlers := &storageHandlers{tv: r.tv, acts: acts}

	group := r.engine.Group("/storages", httputil.RequireAdminRole(errors.ErrAdminRequired))

	// swagger:operation POST /storages Storages CreateStorage
	//
	// Create storage.
	//
	// ---
	// parameters:
	//  - $ref: '#/parameters/UserIDHeader'
	//  - $ref: '#/parameters/UserRoleHeader'
	//  - $ref: '#/parameters/SubstitutedUserID'
	//  - name: body
	//    in: body
	//    required: true
	//    schema:
	//      $ref: '#/definitions/Storage'
	// responses:
	//   '201':
	//     description: storage created
	//   default:
	//     $ref: '#/responses/error'
	group.POST("", handlers.createStorageHandler)

	// swagger:operation GET /storages Storages GetStorages
	//
	// Get storage list.
	//
	// ---
	// parameters:
	//  - $ref: '#/parameters/UserIDHeader'
	//  - $ref: '#/parameters/UserRoleHeader'
	//  - $ref: '#/parameters/SubstitutedUserID'
	// responses:
	//   '200':
	//     description: storages list
	//     schema:
	//       type: array
	//       items:
	//         $ref: '#/definitions/Storage'
	//   default:
	//     $ref: '#/responses/error'
	group.GET("", handlers.getStoragesHandler)

	// swagger:operation PUT /storages/{name} Storages UpdateStorage
	//
	// Update storage.
	//
	// ---
	// parameters:
	//  - $ref: '#/parameters/UserIDHeader'
	//  - $ref: '#/parameters/UserRoleHeader'
	//  - $ref: '#/parameters/SubstitutedUserID'
	//  - name: body
	//    in: body
	//    required: true
	//    schema:
	//      $ref: '#/definitions/UpdateStorageRequest'
	//  - name: name
	//    in: path
	//    type: string
	//    required: true
	// responses:
	//   '202':
	//     description: storage updated
	//   default:
	//     $ref: '#/responses/error'
	group.PUT("/:name", handlers.updateStorageHandler)

	// swagger:operation DELETE /storages/{name} Storages DeleteStorage
	//
	// Delete storage.
	//
	// ---
	// parameters:
	//  - $ref: '#/parameters/UserIDHeader'
	//  - $ref: '#/parameters/UserRoleHeader'
	//  - $ref: '#/parameters/SubstitutedUserID'
	//  - name: name
	//    in: path
	//    type: string
	//    required: true
	// responses:
	//   '202':
	//     description: storage deleted
	//   default:
	//     $ref: '#/responses/error'
	group.DELETE("/:name", handlers.deleteStorageHandler)

	// swagger:operation POST /import/storages Storages ImportStorages
	//
	// Import storages.
	//
	// ---
	// parameters:
	//  - $ref: '#/parameters/UserIDHeader'
	//  - $ref: '#/parameters/UserRoleHeader'
	// responses:
	//   '202':
	//     description: storages imported
	//     schema:
	//       $ref: '#/definitions/ImportResponse'
	//   default:
	//     $ref: '#/responses/error'
	r.engine.POST("/import/storages", handlers.importStoragesHandler)
}
