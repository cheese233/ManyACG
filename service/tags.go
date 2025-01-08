package service

import (
	"context"
	"errors"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/krau/ManyACG/dao"
	"github.com/krau/ManyACG/dao/collections"
	"github.com/krau/ManyACG/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func GetRandomTags(ctx context.Context, limit int) ([]string, error) {
	tags, err := dao.GetRandomTags(ctx, limit)
	if err != nil {
		return nil, err
	}
	tagNames := make([]string, 0, len(tags))
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}
	return tagNames, nil
}

func GetRandomTagModels(ctx context.Context, limit int) ([]*types.TagModel, error) {
	tags, err := dao.GetRandomTags(ctx, limit)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func GetTagByName(ctx context.Context, name string) (*types.TagModel, error) {
	return dao.GetTagByName(ctx, name)
}

// 为已有 tag 添加别名
//
// 同时检查是否有其他 tag 的 name 为所指定的别名之一, 在添加完成后, 删除这些 tag, 并将其对应的 artwork 指向新的 tag (即传入的tagID)
func AddTagAliasByID(ctx context.Context, tagID primitive.ObjectID, alias ...string) (*types.TagModel, error) {
	tagModel, err := dao.GetTagByID(ctx, tagID)
	if err != nil {
		return nil, err
	}

	session, err := dao.Client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	if tagModel.Alias == nil {
		tagModel.Alias = make([]string, 0)
	}
	tagAlias := slice.Union(slice.Concat(tagModel.Alias, alias))

	result, err := session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		if err := dao.UpdateTagAliasByID(ctx, tagID, tagAlias); err != nil {
			return nil, err
		}

		for _, alias := range tagAlias {
			if alias == tagModel.Name {
				continue
			}
			aliasTag, err := dao.GetTagByName(ctx, alias)
			if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
				return nil, err
			}
			if aliasTag == nil {
				continue
			}
			// 添加 aliasTag 中原 alias
			if len(aliasTag.Alias) > 0 {
				regetTagModel, err := dao.GetTagByID(ctx, tagID)
				if err != nil {
					return nil, err
				}
				tagAliasWithaliasTagOriginAlias := slice.Union(slice.Concat(regetTagModel.Alias, aliasTag.Alias))
				if err := dao.UpdateTagAliasByID(ctx, tagID, tagAliasWithaliasTagOriginAlias); err != nil {
					return nil, err
				}
			}

			// 迁移 artwork 中的 tag
			artworkCollection := dao.GetCollection(ctx, collections.Artworks)

			filter := bson.M{"tags": bson.M{"$in": []primitive.ObjectID{aliasTag.ID}}}
			cursor, err := artworkCollection.Find(ctx, filter, options.Find().SetProjection(bson.M{"_id": 1}))
			if err != nil {
				return nil, err
			}
			var matchedDocs []bson.M
			if err = cursor.All(ctx, &matchedDocs); err != nil {
				return nil, err
			}

			if len(matchedDocs) > 0 {
				var docIDs []primitive.ObjectID
				for _, doc := range matchedDocs {
					docIDs = append(docIDs, doc["_id"].(primitive.ObjectID))
				}

				_, err = artworkCollection.UpdateMany(ctx,
					bson.M{"_id": bson.M{"$in": docIDs}},
					bson.M{
						"$pull": bson.M{"tags": aliasTag.ID},
					},
				)
				if err != nil {
					return nil, err
				}

				_, err = artworkCollection.UpdateMany(ctx,
					bson.M{"_id": bson.M{"$in": docIDs}},
					bson.M{
						"$addToSet": bson.M{"tags": tagID},
					},
				)
				if err != nil {
					return nil, err
				}
			}

			// 删除别名对应的 tag
			if _, err := dao.DeleteTagByID(ctx, aliasTag.ID); err != nil {
				return nil, err
			}
		}

		tagModel, err := dao.GetTagByID(ctx, tagID)
		if err != nil {
			return nil, err
		}
		return tagModel, nil
	}, options.Transaction().SetReadPreference(readpref.Primary()))
	if err != nil {
		return nil, err
	}
	return result.(*types.TagModel), nil
}
