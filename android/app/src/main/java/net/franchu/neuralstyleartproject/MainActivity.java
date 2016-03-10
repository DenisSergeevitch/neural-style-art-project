package net.franchu.neuralstyleartproject;

import android.content.Intent;
import android.graphics.Bitmap;
import android.graphics.BitmapFactory;
import android.net.Uri;
import android.os.Environment;
import android.provider.MediaStore;
import android.support.v7.app.AppCompatActivity;
import android.os.Bundle;
import android.util.Log;
import android.view.Gravity;
import android.view.View;
import android.widget.Button;
import android.widget.EditText;
import android.widget.ImageView;
import android.widget.Toast;

import java.io.ByteArrayOutputStream;
import java.io.File;
import java.net.ConnectException;
import java.text.SimpleDateFormat;
import java.util.Date;

import io.grpc.StatusRuntimeException;


public class MainActivity extends AppCompatActivity {
    public static final int MEDIA_TYPE_IMAGE = 1;

    private static final int CAPTURE_IMAGE_ACTIVITY_REQUEST_CODE = 100;
    private Uri fileUri;

    private ImagerClient client;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        Button takePhotoButton = (Button) findViewById(R.id.photoButton);
        takePhotoButton.setOnClickListener(new TakePhotoClickListener());

        Button clearButton = (Button) findViewById(R.id.clearButton);
        clearButton.setOnClickListener(new ClearClickListener());

        Button sendButton = (Button) findViewById(R.id.sendButton);
        sendButton.setOnClickListener(new SendClickListener());

        EditText nameView = (EditText)findViewById(R.id.photoName);
        nameView.setOnClickListener(new PhotoNameClickListener());
        nameView.setOnFocusChangeListener(new PhotoNameFocusChangeListener());

//        client = new ImagerClient("192.168.3.7", 8081);
        client = new ImagerClient("10.9.0.1", 8081);
    }

    public class TakePhotoClickListener implements View.OnClickListener {
        public void onClick(View v) {
            //process event
            // create Intent to take a picture and return control to the calling application
            Intent intent = new Intent(MediaStore.ACTION_IMAGE_CAPTURE);

            fileUri = getOutputMediaFileUri(MEDIA_TYPE_IMAGE); // create a file to save the image
            intent.putExtra(MediaStore.EXTRA_OUTPUT, fileUri); // set the image file name

            // start the image capture Intent
            if (intent.resolveActivity(getPackageManager()) != null) {
                startActivityForResult(intent, CAPTURE_IMAGE_ACTIVITY_REQUEST_CODE);
            }
        }
    }

    public class ClearClickListener implements View.OnClickListener {
        public void onClick(View v) {
            //process event
            ImageView imgView = (ImageView) findViewById(R.id.imageView);
            Bitmap b = Bitmap.createBitmap(imgView.getWidth(), imgView.getHeight(), Bitmap.Config.ALPHA_8);
            imgView.setImageBitmap(b);

            EditText nameView = (EditText) findViewById(R.id.photoName);
            nameView.setText("Name");

            fileUri = null;
        }
    }

    public class SendClickListener implements View.OnClickListener {
        public void onClick(View v) {
            //process event
            if (fileUri == null) {
                Toast.makeText(getApplicationContext(), "No image available", Toast.LENGTH_SHORT).show();
                return;
            }
            EditText nameView = (EditText) findViewById(R.id.photoName);

            Bitmap bitmap = getPhotoBitmap(fileUri, 512, 512);

            ByteArrayOutputStream outStream = new ByteArrayOutputStream();

            bitmap.compress(Bitmap.CompressFormat.JPEG, 95, outStream);

            try {
                client.createJob(nameView.getText().toString(), outStream.toByteArray());
            } catch (StatusRuntimeException e) {
                Toast.makeText(getApplicationContext(), "Server not available", Toast.LENGTH_SHORT).show();
                return;
            }

            Toast.makeText(getApplicationContext(), "Processing images in the server", Toast.LENGTH_SHORT).show();

            Button clearButton = (Button) findViewById(R.id.clearButton);
            clearButton.callOnClick();

        }
    }

    public class PhotoNameClickListener implements View.OnClickListener {
        public void onClick(View v) {
            //process event
            EditText nameView = (EditText) v;
            if (nameView.getText().toString().equals("Name")) {
                nameView.setText("");
            }
        }
    }

    public class PhotoNameFocusChangeListener implements View.OnFocusChangeListener {
        @Override
        public void onFocusChange(View v, boolean hasFocus) {
            if (!hasFocus) {
                EditText nameView = (EditText) v;
                if (nameView.getText().toString().trim().isEmpty()) {
                    nameView.setText("Name");
                }
            }
        }
    }

    @Override
    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        if (requestCode == CAPTURE_IMAGE_ACTIVITY_REQUEST_CODE && resultCode == RESULT_OK) {
            ImageView imgView = (ImageView) findViewById(R.id.imageView);

            Bitmap imageBitmap;
            if (data != null) {
                Bundle extras = data.getExtras();
                imageBitmap = (Bitmap) extras.get("data");

            } else {
                imageBitmap = getPhotoBitmap(fileUri, imgView.getWidth(), imgView.getHeight());
            }
            imgView.setImageBitmap(imageBitmap);
        }
    }

    private Bitmap getPhotoBitmap(Uri uri, int targetWidth, int targetHeight) {
        // Get the dimensions of the bitmap
        BitmapFactory.Options bmOptions = new BitmapFactory.Options();
        bmOptions.inJustDecodeBounds = true;
        BitmapFactory.decodeFile(uri.getPath(), bmOptions);
        int photoW = bmOptions.outWidth;
        int photoH = bmOptions.outHeight;

        // Determine how much to scale down the image
        int scaleFactor = Math.min(photoW/targetWidth, photoH/targetHeight);

        // Decode the image file into a Bitmap sized to fill the View
        bmOptions.inJustDecodeBounds = false;
        bmOptions.inSampleSize = scaleFactor;
        bmOptions.inPurgeable = true;

        return BitmapFactory.decodeFile(uri.getPath(), bmOptions);
    }

    @Override
    protected void onSaveInstanceState(Bundle outState) {
        super.onSaveInstanceState(outState);

        // save file url in bundle as it will be null on screen orientation
        // changes
        outState.putParcelable("file_uri", fileUri);
    }

    @Override
    protected void onRestoreInstanceState(Bundle savedInstanceState) {
        super.onRestoreInstanceState(savedInstanceState);

        // get the file url
        fileUri = savedInstanceState.getParcelable("file_uri");
    }

    private static Uri getOutputMediaFileUri(int type) {
        return Uri.fromFile(getOutputMediaFile(type));
    }

    /**
     * Create a File for saving an image or video
     */
    private static File getOutputMediaFile(int type) {
        // To be safe, you should check that the SDCard is mounted
        // using Environment.getExternalStorageState() before doing this.

        File mediaStorageDir = new File(Environment.getExternalStoragePublicDirectory(
                Environment.DIRECTORY_PICTURES), "neural-style-art-project");
        // This location works best if you want the created images to be shared
        // between applications and persist after your app has been uninstalled.

        // Create the storage directory if it does not exist
        if (!mediaStorageDir.exists()) {
            if (!mediaStorageDir.mkdirs()) {
                Log.d("neural-style-artproject", "failed to create directory");
                return null;
            }
        }

        // Create a media file name
        String timeStamp = new SimpleDateFormat("yyyyMMdd_HHmmss").format(new Date());
        File mediaFile;
        if (type == MEDIA_TYPE_IMAGE) {
            mediaFile = new File(mediaStorageDir.getPath() + File.separator +
                    "IMG_" + timeStamp + ".jpg");
        } else {
            return null;
        }

        return mediaFile;
    }

}
