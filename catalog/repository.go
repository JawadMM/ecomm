package catalog

import (
	"context"
	"encoding/json"
	"log"

	"github.com/olivere/elastic/v7"
)

type Respository interface {
	Close()
	PutProduct(ctx context.Context, product *Product) (*Product, error)
	GetProduct(ctx context.Context, id string) (*Product, error)
	ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error)
	ListProductsWithIds(ctx context.Context, ids []string) ([]Product, error)
	SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error)
}

type elasticRepository struct {
	client *elastic.Client
}

type productDocument struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

func NewElasticRespository(url string) (*elasticRepository, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(url),
		elastic.SetSniff(false),
	)
	if err != nil {
		return nil, err
	}
	return &elasticRepository{client: client}, nil
}

func (r *elasticRepository) Close() {
	r.client.Stop()
}

func (r *elasticRepository) PutProduct(ctx context.Context, product *Product) (*Product, error) {
	_, err := r.client.Index().
		Index("product").
		Id(product.ID).
		BodyJson(&productDocument{
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
		}).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (r *elasticRepository) GetProduct(ctx context.Context, id string) (*Product, error) {
	res, err := r.client.Get().
		Index("product").
		Id(id).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	if !res.Found {
		return nil, nil
	}
	doc := productDocument{}
	if err = json.Unmarshal(res.Source, &doc); err != nil {
		return nil, err
	}

	return &Product{
		ID:          id,
		Name:        doc.Name,
		Description: doc.Description,
		Price:       doc.Price,
	}, nil
}

func (r *elasticRepository) ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error) {
	res, err := r.client.Search().
		Index("product").
		From(int(skip)).
		Size(int(take)).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	var products []Product
	for _, hit := range res.Hits.Hits {
		doc := productDocument{}
		if err = json.Unmarshal(hit.Source, &doc); err == nil {
			products = append(products, Product{
				ID:          hit.Id,
				Name:        doc.Name,
				Description: doc.Description,
				Price:       doc.Price,
			})
		}

	}
	return products, err
}

func (r *elasticRepository) ListProductsWithIds(ctx context.Context, ids []string) ([]Product, error) {
	items := make([]*elastic.MultiGetItem, len(ids))
	for i, id := range ids {
		items[i] = elastic.NewMultiGetItem().Index("product").Id(id)
	}

	res, err := r.client.MultiGet().
		Add(items...).
		Do(ctx)

	if err != nil {
		log.Println("Error fetching products with ids:", err)
		return nil, err
	}

	products := make([]Product, 0, len(ids))
	for _, item := range res.Docs {
		if item.Found {
			doc := productDocument{}
			if err = json.Unmarshal(item.Source, &doc); err == nil {
				products = append(products, Product{
					ID:          item.Id,
					Name:        doc.Name,
					Description: doc.Description,
					Price:       doc.Price,
				})
			}
		}
	}

	return products, nil
}

func (r *elasticRepository) SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error) {
	res, err := r.client.Search().
		Index("product").
		Query(elastic.NewMultiMatchQuery(query, "name", "description")).
		From(int(skip)).
		Size(int(take)).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	products := []Product{}
	for _, hit := range res.Hits.Hits {
		doc := productDocument{}
		if err = json.Unmarshal(hit.Source, &doc); err == nil {
			products = append(products, Product{
				ID:          hit.Id,
				Name:        doc.Name,
				Description: doc.Description,
				Price:       doc.Price,
			})
		}
	}
	return products, nil
}
