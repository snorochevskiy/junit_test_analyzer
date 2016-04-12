package main

import (
	"encoding/json"
	"net/http"
)

type BranchStatusDto struct {
	BranchName string `json:"branch_name"`
	Status     string `json:"last_build_status"`
	Date       string `json:"date"`
}

func serveApiBranchesStatus(context *HttpContext) {
	branches, daoErr := DAO.GetAllBranchesInfo(nil)
	if daoErr != nil {
		http.Error(context.Resp, daoErr.Error(), http.StatusInternalServerError)
		return
	}

	dtos := make([]*BranchStatusDto, 0, len(branches))
	for _, v := range branches {
		dto := new(BranchStatusDto)
		dto.BranchName = v.BranchName
		if v.LastLauchFailed {
			dto.Status = "FAILED"
		} else {
			dto.Status = "HEALTHY"
		}

		dto.Date = v.CreationDate.Format("2006-01-02 15:04:05")
		dtos = append(dtos, dto)
	}

	response, jsonErr := json.Marshal(dtos)
	if jsonErr != nil {
		http.Error(context.Resp, jsonErr.Error(), http.StatusInternalServerError)
		return
	}

	context.Resp.Header().Set("Content-Type", "text/json")
	context.Resp.Write(response)
}
