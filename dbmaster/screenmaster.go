package dbmaster

import (
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/rohit123sinha456/digitalSignage/common"
	DataModel "github.com/rohit123sinha456/digitalSignage/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func checkifcontentlistexists(ctx context.Context, coll *mongo.Collection, screenblocks []DataModel.ScreenBlock) error {
	log.Printf("Checking Contents")
	var objectIDs []primitive.ObjectID

	for _, screenblock := range screenblocks {
		log.Printf("%s", screenblock.BlockName)
		contentlistId := screenblock.ContentListID
		objectIDs = append(objectIDs, contentlistId)
	}
	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return err
	}
	log.Printf("Count is %d", count)
	if len(objectIDs) != int(count) {
		return errors.New("Count Mismatch Error")
	}
	return nil
}

func CreateScreen(ctx context.Context, client *mongo.Client, userID string, screendetails DataModel.Screen) (string, error) {

	screendetails.ID = primitive.NewObjectID()
	userdBname := common.ExtractUserSystemIdentifier(userID)
	coll := client.Database(userdBname).Collection("screen")
	result, inserterr := coll.InsertOne(ctx, screendetails)
	if inserterr != nil {
		return "", inserterr
	}
	log.Printf("Screen Created")
	idstring := result.InsertedID.(primitive.ObjectID).Hex()
	return idstring, nil
}

func ReadScreen(ctx context.Context, client *mongo.Client, userID string) ([]DataModel.Screen, error) {
	var contentlistarray []DataModel.Screen
	userSystemname := common.ExtractUserSystemIdentifier(userID)
	coll := client.Database(userSystemname).Collection("screen")
	filter := bson.D{{}}
	curr, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	for curr.Next(ctx) {
		var result DataModel.Screen
		if err := curr.Decode(&result); err != nil {
			return nil, err
		}
		contentlistarray = append(contentlistarray, result)
	}
	defer curr.Close(ctx)
	return contentlistarray, nil
}

func ReadOneScreen(ctx context.Context, client *mongo.Client, userID string, screenID string) (DataModel.Screen, error) {
	var result DataModel.Screen
	userSystemname := common.ExtractUserSystemIdentifier(userID)
	coll := client.Database(userSystemname).Collection("screen")
	objectId, err := primitive.ObjectIDFromHex(screenID)
	if err != nil {
		return result, err
	}
	filter := bson.D{{"_id", objectId}}
	err = coll.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func UpdateScreen(ctx context.Context, client *mongo.Client, userID string, screenID string, screenblock []DataModel.ScreenBlock) error {
	userSystemname := common.ExtractUserSystemIdentifier(userID)
	coll := client.Database(userSystemname).Collection("screen")
	userdBname := common.ExtractUserSystemIdentifier(userID)
	contentListCollection := client.Database(userdBname).Collection("contentlist")
	err := checkifcontentlistexists(ctx, contentListCollection, screenblock)
	if err != nil {
		return err
	}
	objectId, err := primitive.ObjectIDFromHex(screenID)
	if err != nil {
		return err
	}
	filter := bson.D{{"_id", objectId}}
	update := bson.D{{"$set", bson.D{{"screenblock", screenblock}}}}

	// Updates the first document that has the specified "_id" value
	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	log.Printf("Documents updated: %v\n", result.ModifiedCount)
	return nil
}

func DeleteScreen(ctx context.Context, client *mongo.Client, userID string, screenID string) error {
	userSystemname := common.ExtractUserSystemIdentifier(userID)
	screenCollection := client.Database(userSystemname).Collection("screen")
	userdBname := common.ExtractUserSystemIdentifier(userID)
	plalistCollection := client.Database(userdBname).Collection("plalist")

	objectId, err := primitive.ObjectIDFromHex(screenID)
	if err != nil {
		return err
	}
	filter := bson.D{{"_id", objectId}}
	result, err := screenCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	log.Printf("Number of documents deleted from Screens: %d\n", result.DeletedCount)
	// Create the update filter and update document
    playlistfilter := bson.M{}
    playlistupdate := bson.M{
        "$pull": bson.M{
            "deviceblock": bson.M{
                "deviceid": objectId,
            },
        },
    }

    // Perform the update operation
    playlistdeleteresult, err := plalistCollection.UpdateMany(ctx, playlistfilter, playlistupdate)
    if err != nil {
        log.Fatal(err)
    }
	log.Printf("Number of documents deleted from Plalist: %d\n", playlistdeleteresult.ModifiedCount)
	return nil
}