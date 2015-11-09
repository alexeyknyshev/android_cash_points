package net.agnia.cashpoints;

import android.location.LocationManager;
import android.content.Intent;
import android.content.Context;

public class CashPointsActivity extends org.qtproject.qt5.android.bindings.QtActivity
{
    private static CashPointsActivity mSingleton;

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
}
