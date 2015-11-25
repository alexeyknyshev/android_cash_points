#include "locationservice.h"

#ifdef Q_OS_ANDROID
#include <QtAndroidExtras/QAndroidJniObject>
#endif // Q_OS_ANDROID

#include <QtPositioning/QGeoRectangle>
#include <QtPositioning/QGeoCircle>

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

qreal LocationService::getGeoRegionRadius(const QGeoShape &shape) const
{
    if (!shape.isValid()) {
        return 0.0;
    }

    const QGeoShape::ShapeType type = shape.type();
    if (type == QGeoShape::RectangleType) {
        const QGeoRectangle &rect = static_cast<const QGeoRectangle &>(shape);
        return rect.center().distanceTo(rect.topLeft());
    } else if (type == QGeoShape::CircleType) {
        const QGeoCircle &circle = static_cast<const QGeoCircle &>(shape);
        return circle.radius();
    }

    return 0.0;
}


