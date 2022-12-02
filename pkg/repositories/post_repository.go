package repositories

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golek_posts_service/pkg/contracts"
	"golek_posts_service/pkg/contracts/status"
	"golek_posts_service/pkg/models"
	"log"
)

type DatabaseRepository struct {
	Connection *mongo.Database
	Collection *mongo.Collection
}

func (d DatabaseRepository) Search(ctx context.Context, keyword string, limit int64, skip int64) ([]models.Post, error) {

	opts := options.Find()
	opts.SetLimit(limit)
	opts.SetSkip(skip)

	filter := bson.D{{"$text", bson.D{{"$search", keyword}}}}
	cursor, err := d.Collection.Find(ctx, filter, opts)
	if err != nil {
		return []models.Post{}, err
	}

	results := make([]models.Post, 0)
	if err = cursor.All(context.TODO(), &results); err != nil {
		return []models.Post{}, err
	}

	return results, nil
}

func (d DatabaseRepository) Fetch(ctx context.Context, latest bool, limit int64, skip int64) ([]models.Post, error) {

	opts := options.Find()
	opts.SetLimit(limit)
	opts.SetSkip(skip)

	if latest {
		opts.SetSort(bson.D{{"created_at", -1}})
	}

	//Fetch Records
	filter := map[string]interface{}{"deleted_at": nil}
	records, err := d.Collection.Find(ctx, filter, opts)
	if err != nil {
		log.Println("DBRepository Error: Fetch")
		return nil, err
	}

	//Close Cursor
	defer func(records *mongo.Cursor, ctx context.Context) {
		err := records.Close(ctx)
		if err != nil {
			panic(err)
		}
	}(records, ctx)

	posts := make([]models.Post, 0)
	//Append Each Record to results
	for records.Next(ctx) {

		var post models.Post

		err := records.Decode(&post)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func (d DatabaseRepository) FindById(ctx context.Context, postID string) (models.Post, error) {

	var post models.Post

	//Convert PostId to ObjectID
	objectID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return post, err
	}

	//Set filters
	filter := map[string]interface{}{"_id": objectID, "deleted_at": nil}

	//Decode
	err = d.Collection.FindOne(ctx, filter).Decode(&post)
	if err != nil {
		return post, err
	}

	return post, nil
}

func (d DatabaseRepository) Create(ctx context.Context, post models.Post) (models.Post, status.PostOperationStatus, error) {

	//Open transaction
	err := d.Connection.Client().UseSession(ctx, func(sessionContext mongo.SessionContext) error {

		//Start Transaction
		err := sessionContext.StartTransaction()
		if err != nil {
			return err
		}

		//Insert Data
		createdPost, err := d.Collection.InsertOne(ctx, post)
		if err != nil {
			return err
		}

		//attach id
		post.SetPostID(createdPost.InsertedID.(primitive.ObjectID))

		//Commit Transaction
		err = sessionContext.CommitTransaction(ctx)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return models.Post{}, status.PostCreatedStatusFailed, err
	}

	return post, status.PostCreatedStatusSuccess, nil
}

func (d DatabaseRepository) Update(ctx context.Context, postID string, post models.Post) (models.Post, status.PostOperationStatus, error) {

	//Convert PostID to Mongo ObjectID
	objectId, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return models.Post{}, status.PostUpdatedStatusFailed, err
	}

	//Query filter
	filter := bson.D{{"_id", objectId}}

	//Query
	_, err = d.Collection.UpdateOne(ctx, filter, bson.D{{"$set", post}})
	if err != nil {
		return models.Post{}, status.PostUpdatedStatusFailed, err
	}
	return post, status.PostUpdatedStatusSuccess, err
}

func (d DatabaseRepository) Delete(ctx context.Context, postID string) (status.PostOperationStatus, error) {

	//Convert PostID to Mongo ObjectID
	objectID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return status.PostDeletedStatusFailed, err
	}

	//Query filter
	filter := bson.D{{"_id", objectID}}

	//Query
	deleteResult, err := d.Collection.DeleteOne(ctx, filter)
	if err != nil {
		return status.PostDeletedStatusFailed, err
	}

	if deleteResult.DeletedCount == 0 {
		return status.PostDeletedStatusFailed, errors.New("unmatched any documents")
	}

	return status.PostDeletedStatusSuccess, nil
}

func NewPostRepository(connection *mongo.Database, collection *mongo.Collection) contracts.PostRepositoryContract {
	return &DatabaseRepository{Connection: connection, Collection: collection}
}
