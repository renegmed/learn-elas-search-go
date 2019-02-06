package util

import (
	"context"

	"github.com/olivere/elastic"
)

func Destroy(index string) error {
	ctx := context.Background()

	// Create a elasticsearch client
	client, err := elastic.NewClient()
	if err != nil {
		return err
	}

	exists, err := client.IndexExists(index).Do(ctx)
	if err != nil {
		return err
	}

	if !exists {
		// Delete an index.
		deleteIndex, err := client.DeleteIndex(index).Do(ctx)
		if err != nil {
			check(err)
		}
		if !deleteIndex.Acknowledged {
			// Not acknowledged
		}
	}

	return nil
}
