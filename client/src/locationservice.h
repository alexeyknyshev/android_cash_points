#ifndef LOCATIONSERVICE_H
#define LOCATIONSERVICE_H

#include <QtCore/QObject>
#include <QtPositioning/QGeoCoordinate>

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

signals:
    void enabledChanged(bool enabled);
    void coordinateChanged(const QGeoCoordinate &coord);

private:
    QGeoCoordinate mLastCoord;
};

#endif // LOCATIONSERVICE_H
