package convenience

import (
	"github.com/calindra/nonodo/internal/convenience/decoder"
	"github.com/calindra/nonodo/internal/convenience/repository"
	"github.com/calindra/nonodo/internal/convenience/services"
	"github.com/calindra/nonodo/internal/convenience/synchronizer"
	"github.com/jmoiron/sqlx"
)

// what is the best DI/IoC framework for go?

type Container struct {
	db                  *sqlx.DB
	outputDecoder       *decoder.OutputDecoder
	convenienceService  *services.ConvenienceService
	repository          *repository.VoucherRepository
	syncRepository      *repository.SynchronizerRepository
	graphQLSynchronizer *synchronizer.Synchronizer
	voucherFetcher      *synchronizer.VoucherFetcher
	noticeRepository    *repository.NoticeRepository
}

func NewContainer(db sqlx.DB) *Container {
	return &Container{
		db: &db,
	}
}

func (c *Container) GetOutputDecoder() *decoder.OutputDecoder {
	if c.outputDecoder != nil {
		return c.outputDecoder
	}
	c.outputDecoder = decoder.NewOutputDecoder(*c.GetConvenienceService())
	return c.outputDecoder
}

func (c *Container) GetRepository() *repository.VoucherRepository {
	if c.repository != nil {
		return c.repository
	}
	c.repository = &repository.VoucherRepository{
		Db: *c.db,
	}
	err := c.repository.CreateTables()
	if err != nil {
		panic(err)
	}
	return c.repository
}

func (c *Container) GetSyncRepository() *repository.SynchronizerRepository {
	if c.syncRepository != nil {
		return c.syncRepository
	}
	c.syncRepository = &repository.SynchronizerRepository{
		Db: *c.db,
	}
	err := c.syncRepository.CreateTables()
	if err != nil {
		panic(err)
	}
	return c.syncRepository
}

func (c *Container) GetNoticeRepository() *repository.NoticeRepository {
	if c.syncRepository != nil {
		return c.noticeRepository
	}
	c.noticeRepository = &repository.NoticeRepository{
		Db: *c.db,
	}
	err := c.noticeRepository.CreateTables()
	if err != nil {
		panic(err)
	}
	return c.noticeRepository
}

func (c *Container) GetConvenienceService() *services.ConvenienceService {
	if c.convenienceService != nil {
		return c.convenienceService
	}
	c.convenienceService = services.NewConvenienceService(
		c.GetRepository(),
		c.GetNoticeRepository(),
	)
	return c.convenienceService
}

func (c *Container) GetGraphQLSynchronizer() *synchronizer.Synchronizer {
	if c.graphQLSynchronizer != nil {
		return c.graphQLSynchronizer
	}
	c.graphQLSynchronizer = synchronizer.NewSynchronizer(
		c.GetOutputDecoder(),
		c.GetVoucherFetcher(),
		c.GetSyncRepository(),
	)
	return c.graphQLSynchronizer
}
func (c *Container) GetVoucherFetcher() *synchronizer.VoucherFetcher {
	if c.voucherFetcher != nil {
		return c.voucherFetcher
	}
	c.voucherFetcher = synchronizer.NewVoucherFetcher()
	return c.voucherFetcher
}
