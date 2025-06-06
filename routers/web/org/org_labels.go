// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package org

import (
	"errors"

	"code.gitea.io/gitea/models/db"
	issues_model "code.gitea.io/gitea/models/issues"
	"code.gitea.io/gitea/modules/label"
	repo_module "code.gitea.io/gitea/modules/repository"
	"code.gitea.io/gitea/modules/util"
	"code.gitea.io/gitea/modules/web"
	shared_label "code.gitea.io/gitea/routers/web/shared/label"
	"code.gitea.io/gitea/services/context"
	"code.gitea.io/gitea/services/forms"
)

// RetrieveLabels find all the labels of an organization
func RetrieveLabels(ctx *context.Context) {
	labels, err := issues_model.GetLabelsByOrgID(ctx, ctx.Org.Organization.ID, ctx.FormString("sort"), db.ListOptions{})
	if err != nil {
		ctx.ServerError("RetrieveLabels.GetLabels", err)
		return
	}
	for _, l := range labels {
		l.CalOpenIssues()
	}
	ctx.Data["Labels"] = labels
	ctx.Data["NumLabels"] = len(labels)
	ctx.Data["SortType"] = ctx.FormString("sort")
}

// NewLabel create new label for organization
func NewLabel(ctx *context.Context) {
	form := shared_label.GetLabelEditForm(ctx)
	if ctx.Written() {
		return
	}

	l := &issues_model.Label{
		OrgID:          ctx.Org.Organization.ID,
		Name:           form.Title,
		Exclusive:      form.Exclusive,
		Description:    form.Description,
		Color:          form.Color,
		ExclusiveOrder: form.ExclusiveOrder,
	}
	if err := issues_model.NewLabel(ctx, l); err != nil {
		ctx.ServerError("NewLabel", err)
		return
	}
	ctx.JSONRedirect(ctx.Org.OrgLink + "/settings/labels")
}

// UpdateLabel update a label's name and color
func UpdateLabel(ctx *context.Context) {
	form := shared_label.GetLabelEditForm(ctx)
	if ctx.Written() {
		return
	}

	l, err := issues_model.GetLabelInOrgByID(ctx, ctx.Org.Organization.ID, form.ID)
	if errors.Is(err, util.ErrNotExist) {
		ctx.JSONErrorNotFound()
		return
	} else if err != nil {
		ctx.ServerError("GetLabelInOrgByID", err)
		return
	}

	l.Name = form.Title
	l.Exclusive = form.Exclusive
	l.ExclusiveOrder = form.ExclusiveOrder
	l.Description = form.Description
	l.Color = form.Color
	l.SetArchived(form.IsArchived)
	if err := issues_model.UpdateLabel(ctx, l); err != nil {
		ctx.ServerError("UpdateLabel", err)
		return
	}
	ctx.JSONRedirect(ctx.Org.OrgLink + "/settings/labels")
}

// DeleteLabel delete a label
func DeleteLabel(ctx *context.Context) {
	if err := issues_model.DeleteLabel(ctx, ctx.Org.Organization.ID, ctx.FormInt64("id")); err != nil {
		ctx.Flash.Error("DeleteLabel: " + err.Error())
	} else {
		ctx.Flash.Success(ctx.Tr("repo.issues.label_deletion_success"))
	}

	ctx.JSONRedirect(ctx.Org.OrgLink + "/settings/labels")
}

// InitializeLabels init labels for an organization
func InitializeLabels(ctx *context.Context) {
	form := web.GetForm(ctx).(*forms.InitializeLabelsForm)
	if ctx.HasError() {
		ctx.Redirect(ctx.Org.OrgLink + "/labels")
		return
	}

	if err := repo_module.InitializeLabels(ctx, ctx.Org.Organization.ID, form.TemplateName, true); err != nil {
		if label.IsErrTemplateLoad(err) {
			originalErr := err.(label.ErrTemplateLoad).OriginalError
			ctx.Flash.Error(ctx.Tr("repo.issues.label_templates.fail_to_load_file", form.TemplateName, originalErr))
			ctx.Redirect(ctx.Org.OrgLink + "/settings/labels")
			return
		}
		ctx.ServerError("InitializeLabels", err)
		return
	}
	ctx.Redirect(ctx.Org.OrgLink + "/settings/labels")
}
