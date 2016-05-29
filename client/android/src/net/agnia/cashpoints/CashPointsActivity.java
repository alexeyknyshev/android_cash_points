package net.agnia.cashpoints;

import android.location.LocationManager;
import android.content.Intent;
import android.content.Context;
import android.os.Bundle;

import com.google.android.gms.auth.api.Auth;
import com.google.android.gms.auth.api.signin.GoogleSignInAccount;
import com.google.android.gms.auth.api.signin.GoogleSignInOptions;
import com.google.android.gms.auth.api.signin.GoogleSignInResult;
import com.google.android.gms.common.ConnectionResult;
import com.google.android.gms.common.SignInButton;
import com.google.android.gms.common.api.GoogleApiClient;
import com.google.android.gms.common.api.GoogleApiClient.ConnectionCallbacks;
import com.google.android.gms.common.api.GoogleApiClient.OnConnectionFailedListener;
import com.google.android.gms.common.api.OptionalPendingResult;
import com.google.android.gms.common.api.ResultCallback;
import com.google.android.gms.common.api.Status;

public class CashPointsActivity extends org.qtproject.qt5.android.bindings.QtActivity
                                implements ConnectionCallbacks, GoogleApiClient.OnConnectionFailedListener
{
    private static CashPointsActivity mSingleton;
    private GoogleApiClient mGoogleApiClient;

    public CashPointsActivity()
    {
        mSingleton = this;
    }

    public static boolean isLocationServiceEnabled()
    {
        LocationManager locationManager = (LocationManager)mSingleton.getSystemService(Context.LOCATION_SERVICE);
        return locationManager.isProviderEnabled(LocationManager.GPS_PROVIDER);
    }

    public static void setLocationServiceEnabled()
    {
        LocationManager locationManager = (LocationManager)mSingleton.getSystemService(Context.LOCATION_SERVICE);
        if(!locationManager.isProviderEnabled(LocationManager.GPS_PROVIDER)) {
            mSingleton.startActivity(new Intent(android.provider.Settings.ACTION_LOCATION_SOURCE_SETTINGS));
        }
    }

    public static void googleApiConnect()
    {
        mSingleton.mGoogleApiClient.connect();
    }


    @Override
    public void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
/*
        GoogleSignInOptions gso = new GoogleSignInOptions.Builder(GoogleSignInOptions.DEFAULT_SIGN_IN)
                        .requestEmail()
                        .build();

        mGoogleApiClient = new GoogleApiClient.Builder(this)
                        .addConnectionCallbacks(this)
                        .addApi(Auth.GOOGLE_SIGN_IN_API, gso)
                        .build();*/
    }

    @Override
    public void onConnected(Bundle b) {
        System.out.println("GoogleApiClient connected");
    }

    @Override
    public void onConnectionFailed(ConnectionResult result) {
        System.out.println("GoogleApiClient connection failed");
    }

    @Override
    public void onConnectionSuspended(int s) {
        System.out.println("GoogleApiClient connection suspended");
    }
}
