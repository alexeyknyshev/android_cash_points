#include "cashpointinradius.h"

#include "../serverapi.h"
#include "../cashpointsqlmodel.h"

#include <QtCore/QDebug>
#include <QtCore/QJsonObject>
#include <QtCore/QJsonParseError>
#include <QtCore/QJsonArray>

CashPointInRadius::CashPointInRadius(CashPointSqlModel *model)
    : CashPointRequest(model),
      mRadius(1000.0f)
{ }

void CashPointInRadius::sendImpl(ServerApi *api, quint32 leftAttempts)
{
    if (leftAttempts == 0) {
        qDebug() << "CashPointInRadius: no retry attempt left";
        return;
    }

    QJsonObject json;
    json["longitude"] = mCoord.longitude();
    json["latitude"] = mCoord.latitude();
    json["radius"] = mRadius;
    api->sendRequest("/nearby/cashpoints", json,
    [&](ServerApi::RequestStatusCode reqCode, ServerApi::HttpStatusCode httpCode, const QByteArray &data) {
        if (reqCode == ServerApi::RSC_Timeout) {
            emitUpdate(leftAttempts - 1);
            return;
        }

        if (reqCode != ServerApi::RSC_Ok) {
            emitError(ServerApi::requestStatusCodeText(reqCode));
            return;
        }

        if (httpCode != ServerApi::HSC_Ok) {
            qWarning() << "Server request http code: " << httpCode;
            emitUpdate(leftAttempts - 1);
            return;
        }

        QJsonParseError err;
        const QJsonDocument json = QJsonDocument::fromJson(data, &err);
        if (err.error != QJsonParseError::NoError) {
            emitError("CashPointInRadius: server response json parse error: " + err.errorString());
            return;
        }

        if (!json.isObject()) {
            emitError("CashPointInRadius: response is not an object!");
            return;
        }

        const QJsonObject obj = json.object();
        if (!obj["cash_points"].isArray()) {
            emitError("CashPointInRadius: cash_points field is not an array!");
            return;
        }

        const QJsonArray arr = obj["cash_points"].toArray();
        const auto end = arr.constEnd();
        for (auto it = arr.constBegin(); it != end; it++) {
            const QJsonValue &val = *it;
            qDebug() << val.toInt();
        }

        setLastUpdateTime(QDateTime::currentDateTime());
    });
}

void CashPointInRadius::fromJson(const QJsonObject &json)
{
    setRadius(json["radius"].toDouble());
    QGeoCoordinate coord;
    coord.setLatitude(json["latitude"].toDouble());
    coord.setLongitude(json["longitude"].toDouble());
    setCoordinate(coord);
}

void CashPointInRadius::setRadius(qreal radius) {
    if (radius <= 0) {
        qDebug() << "cashpoint search radius must be positive";
        return;
    }
    mRadius = radius;
}

void CashPointInRadius::setCoordinate(const QGeoCoordinate &coord)
{
    if (!coord.isValid()) {
        qDebug() << "cashpoinst search coordinate must be valid";
        return;
    }
    mCoord = coord;
}
