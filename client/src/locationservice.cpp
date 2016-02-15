#include "locationservice.h"

#ifdef Q_OS_ANDROID
#include <QtAndroidExtras/QAndroidJniObject>
#endif // Q_OS_ANDROID

#include <QtPositioning/QGeoRectangle>
#include <QtPositioning/QGeoCircle>

#include <QtCore/QDebug>

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

qreal LocationService::getGeoRegionRadiusEstimate(const QGeoCoordinate &from, const QGeoCoordinate &to) const
{
    const qreal distance = from.distanceTo(to);
    qDebug() << "distance: " << distance;
    return distance;
}
/*{
    if (zoomLevel > 19.0) {
        return 79.0f;
    } else if (zoomLevel < 0.0) {
        return 4929715;
    }

    static const qreal map[] = { 4929715, 4929715, 4929715, 4929715, 2484252, 1252514,
                                 632297,  318220,  159704,  80010,   40046,   20033,
                                 10019,   5010,    2505,    1252,    626,     313,
                                 157,     79,      79 };

    const int index = (int)std::floor(zoomLevel);
    const qreal multiplier = zoomLevel - index;
    const qreal higher = map[index];
    const qreal lower = map[index + 1];
    return lower + ( (higher - lower) * multiplier );
}*/


