#include "googleapiclient.h"

#ifdef Q_OS_ANDROID
#include <QtAndroidExtras/QAndroidJniObject>
#endif // Q_OS_ANDROID

GoogleApiClient::GoogleApiClient(QObject *parent)
    : QObject(parent)
{ }

void GoogleApiClient::dial()
{
#ifdef Q_OS_ANDROID
//    QAndroidJniObject::callStaticMethod<void>("net/agnia/cashpoints/CashPointsActivity",
//                                              "googleApiConnect");
#endif // Q_OS_ANDROID
}
