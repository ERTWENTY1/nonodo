package model

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/calindra/nonodo/internal/commons"
	"github.com/calindra/nonodo/internal/convenience/model"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jmoiron/sqlx"
)

const INPUT_INDEX = "InputIndex"

type ReportRepository struct {
	Db *sqlx.DB
}

func (r *ReportRepository) CreateTables() error {
	schema := `CREATE TABLE IF NOT EXISTS reports (
		output_index	integer,
		payload 		text,
		input_index 	integer);`
	_, err := r.Db.Exec(schema)
	if err == nil {
		slog.Debug("Reports table created")
	} else {
		slog.Error("Create table error", err)
	}
	return err
}

func (r *ReportRepository) Create(report Report) (Report, error) {
	insertSql := `INSERT INTO reports (
		output_index,
		payload,
		input_index) VALUES (?, ?, ?)`
	r.Db.MustExec(
		insertSql,
		report.Index,
		common.Bytes2Hex(report.Payload),
		report.InputIndex,
	)
	return report, nil
}

func (r *ReportRepository) FindByInputAndOutputIndex(
	inputIndex uint64,
	outputIndex uint64,
) (*Report, error) {
	rows, err := r.Db.Queryx(`
		SELECT payload FROM reports 
			WHERE input_index = ? and output_index = ?
			LIMIT 1`,
		inputIndex, outputIndex,
	)
	if err != nil {
		slog.Error("database error", "err", err)
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		report := &Report{
			InputIndex: int(inputIndex),
			Index:      int(outputIndex),
			Payload:    common.Hex2Bytes(payload),
		}
		return report, nil
	} else {
		return nil, nil
	}
}

func (c *ReportRepository) Count(
	filter []*model.ConvenienceFilter,
) (uint64, error) {
	query := `SELECT count(*) FROM reports `
	where, args, err := transformToQuery(filter)
	if err != nil {
		slog.Error("Count execution error")
		return 0, err
	}
	query += where
	slog.Debug("Query", "query", query, "args", args)
	stmt, err := c.Db.Preparex(query)
	if err != nil {
		slog.Error("Count execution error")
		return 0, err
	}
	var count uint64
	err = stmt.Get(&count, args...)
	if err != nil {
		slog.Error("Count execution error")
		return 0, err
	}
	return count, nil
}

func (c *ReportRepository) FindAllByInputIndex(
	first *int,
	last *int,
	after *string,
	before *string,
	inputIndex *int,
) (*commons.PageResult[Report], error) {
	filter := []*model.ConvenienceFilter{}
	if inputIndex != nil {
		field := INPUT_INDEX
		value := fmt.Sprintf("%d", *inputIndex)
		filter = append(filter, &model.ConvenienceFilter{
			Field: &field,
			Eq:    &value,
		})
	}
	return c.FindAll(
		first,
		last,
		after,
		before,
		filter,
	)
}

func (c *ReportRepository) FindAll(
	first *int,
	last *int,
	after *string,
	before *string,
	filter []*model.ConvenienceFilter,
) (*commons.PageResult[Report], error) {
	total, err := c.Count(filter)
	if err != nil {
		slog.Error("database error", "err", err)
		return nil, err
	}
	query := `SELECT input_index, output_index, payload FROM reports `
	where, args, err := transformToQuery(filter)
	if err != nil {
		slog.Error("database error", "err", err)
		return nil, err
	}
	query += where
	query += `ORDER BY input_index ASC, output_index ASC `
	offset, limit, err := commons.ComputePage(first, last, after, before, int(total))
	if err != nil {
		return nil, err
	}
	query += `LIMIT ? `
	args = append(args, limit)
	query += `OFFSET ? `
	args = append(args, offset)

	slog.Debug("Query", "query", query, "args", args, "total", total)
	stmt, err := c.Db.Preparex(query)
	if err != nil {
		return nil, err
	}
	var reports []Report
	rows, err := stmt.Queryx(args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var payload string
		var inputIndex int
		var outputIndex int
		if err := rows.Scan(&inputIndex, &outputIndex, &payload); err != nil {
			return nil, err
		}
		report := &Report{
			InputIndex: inputIndex,
			Index:      outputIndex,
			Payload:    common.Hex2Bytes(payload),
		}
		reports = append(reports, *report)
	}

	pageResult := &commons.PageResult[Report]{
		Rows:   reports,
		Total:  total,
		Offset: uint64(offset),
	}
	return pageResult, nil
}

func transformToQuery(
	filter []*model.ConvenienceFilter,
) (string, []interface{}, error) {
	query := ""
	if len(filter) > 0 {
		query += "WHERE "
	}
	args := []interface{}{}
	where := []string{}
	for _, filter := range filter {
		if *filter.Field == "OutputIndex" {
			if filter.Eq != nil {
				where = append(where, "output_index = ?")
				args = append(args, *filter.Eq)
			} else {
				return "", nil, fmt.Errorf("operation not implemented")
			}
		} else if *filter.Field == INPUT_INDEX {
			if filter.Eq != nil {
				where = append(where, "input_index = ?")
				args = append(args, *filter.Eq)
			} else {
				return "", nil, fmt.Errorf("operation not implemented")
			}
		} else {
			return "", nil, fmt.Errorf("unexpected field %s", *filter.Field)
		}
	}
	query += strings.Join(where, " and ")
	return query, args, nil
}
