#include "locationservice.h"

#ifdef Q_OS_ANDROID
#include <QtAndroidExtras/QAndroidJniObject>
#endif // Q_OS_ANDROID

LocationService::LocationService(QObject *parent)
    : QObject(parent)
{ }

void LocationService::setEnabled(bool enable)
{
#ifdef Q_OS_ANDROID
    Q_UNUSED(enable);
    QAndroidJniObject::callStaticMethod<void>("net/agnia/cashpoints/CashPointsActivity",
                                              "setLocationServiceEnabled");
#endif // Q_OS_ANDROID
    emit enabledChanged(isEnabled());
}

bool LocationService::isEnabled() const
{
#ifdef Q_OS_ANDROID
    const jboolean enabled = QAndroidJniObject::callStaticMethod<jboolean>(
                "net/agnia/cashpoints/CashPointsActivity",
                "isLocationServiceEnabled");
#else
    bool enabled = true;
#endif // Q_OS_ANDROID
    return (bool)enabled;
}

void LocationService::updateCoordinate()
{

}


