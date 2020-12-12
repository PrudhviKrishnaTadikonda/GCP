Uploading Images Through a Signed URL
Objective: 
The main objective of this project is to upload image to Google cloud Storage bucket using a signed URL. 
Google Cloud Functions are triggered to validate the format of the uploaded image, then the Google Vison API creates labels for the images and tell if the upload is appropriate or not.
Finally, if the image is appropriate it is  available for end users through a signed URL.

Cloud Components Used:
1. App Engine 
2. Cloud Storage 
3. Cloud Functions 

Project Workflow:
1. When the app engine receives a request from user it Creates a signed URL that allows the user to upload images to cloud storage. 
2. Image is uploaded to a particular bucket in cloud storage. 
3. The uploaded image is stored in cloud storage bucket and then google cloud function validates the image format and size. 
4. After verifying the image format, the cloud vison API is used to find out if the image is safe or not.
5. Then the image is copied to another bucket in cloud storage 
6. Now, the app engine provides the end user a signed URL to access the image in the distribution bucket.

API’s Used:
1.Cloud Functions API
2. Cloud Storage API 
3. Cloud Vision API
Service Account: It must be give “Storage Admin” and “Token Generator” roles

Steps To Run the Service:
Step 1: Create Buckets in GCP
Set all attributes 
Create Upload bucket and distribution bucket
Publishing all objects to all types of users
 
Step 2: Deploy the Signed URL generation Code To App Engine
Deploy the code present in Image-Censor-AppEngine in github to App Enging using 
gcloud app deploy
The app will be accessible at https://image-censor.ue.r.appspot.com
Post call to https://image-censor.ue.r.appspot.com/sign  will create the signed url for uploading

Step 3: Deploy the Google Cloud Functions
The code is present in IMAGE_CENSOR_APP_FUNCTIONS run the following command
$ UPLOADABLE_BUCKET="my_uploadable_bucket"
gcloud functions deploy UploadImage --runtime go111 --trigger-resource $UPLOADABLE_BUCKET --trigger-event goog
le.storage.object.finalize --retry

Step 4: Upload The Image 
Now run a go program that does a put request to our app and also send an image.
The code is in client folder of git repository.
Output: The signed URL and the image description.
 
