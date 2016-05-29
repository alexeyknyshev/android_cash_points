#ifndef LOCATIONSERVICE_H
#define LOCATIONSERVICE_H

#include <QtCore/QObject>
#include <QtPositioning/QGeoCoordinate>
#include <QtPositioning/QGeoShape>

class LocationService : public QObject
{
    Q_OBJECT

    Q_PROPERTY(bool enabled READ isEnabled WRITE setEnabled NOTIFY enabledChanged)

public:
    explicit LocationService(QObject *parent = 0);

    void setEnabled(bool enable);
    bool isEnabled() const;

    void updateCoordinate();
    const QGeoCoordinate &getCoordinate() const { return mLastCoord; }

    Q_INVOKABLE qreal getGeoRegionRadius(const QGeoShape &shape) const;
    Q_INVOKABLE qreal getGeoRegionRadiusEstimate(const QGeoCoordinate &from, const QGeoCoordinate &to) const;

signals:
    void enabledChanged(bool enabled);
    void coordinateChanged(const QGeoCoordinate &coord);

private:
    QGeoCoordinate mLastCoord;
};

#endif // LOCATIONSERVICE_H
