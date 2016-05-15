#include "nearbyclusters.h"

#include "../serverapi.h"

#include <QtCore/QDebug>
#include <QtCore/QJsonObject>
#include <QtCore/QJsonParseError>
#include <QtCore/QJsonArray>

NearbyClusters::NearbyClusters(CashPointSqlModel *model)
    : CashPointRequest(model)
{
    registerStepHandlers({
                             STEP_HANDLER(fetchClusters)
                         });
}

NearbyClusters::~NearbyClusters()
{
    qDebug() << "destroyed " << this;
}

bool NearbyClusters::fromJson(const QJsonObject &json)
{
    const QJsonValue longitude = json["longitude"];
    const QJsonValue latitude  = json["latitude"];
    const QJsonValue radius    = json["radius"];
    const QJsonValue zoom      = json["zoom"];
    const QJsonValue filter    = json["filter"];

    const QJsonValue topLeft =   json["topLeft"];
    const QJsonValue botRight =  json["bottomRight"];

    if (!zoom.isDouble() || !radius.isDouble() ||
        !longitude.isDouble() || !latitude.isDouble() ||
        !(filter.isObject() || filter.isUndefined()) ||
        !(topLeft.isObject()) || !botRight.isObject() )
    {
        auto it = data.begin();
        while (it != data.end()) {
            it = data.erase(it);
        }
        return false;
    }

    data["longitude"] = longitude;
    data["latitude"] = latitude;
    data["radius"] = radius;
    data["zoom"] = qRound(zoom.toDouble());
    data["topLeft"] = topLeft;
    data["bottomRight"] = botRight;

    if (filter.isObject()) {
        data["filter"] = filter;
    }

    return true;
}

void NearbyClusters::fetchClusters(ServerApiPtr api, quint32 leftAttempts)
{
    qDebug() << "fetchClusters";

    if (data.isEmpty()) {
        return;
    }

    mId = api->sendRequest("/nearby/clusters", data,
    [this, api, leftAttempts]
    (ServerApi::RequestStatusCode reqCode, ServerApi::HttpStatusCode httpCode, const QByteArray &data) {
        qDebug() << "callback " << this;
        if (!isRunning()) {
            if (isDisposing()) {
                deleteLater();
            }
            return;
        }
        const int step = 1;

        const QString errText = trUtf8("Cannot receive data of nearby clusters");
        if (reqCode == ServerApi::RSC_Timeout) {
            emitUpdate(leftAttempts - 1, step);
            return;
        }

        if (reqCode != ServerApi::RSC_Ok) {
            emitError(ServerApi::requestStatusCodeText(reqCode));
            emitStepFinished(*api, step, false, errText);
            return;
        }

        if (httpCode != ServerApi::HSC_Ok) {
            qWarning() << "Server request http code: " << httpCode;
            emitUpdate(leftAttempts - 1, step);
            return;
        }

        QJsonParseError err;
        const QJsonDocument json = QJsonDocument::fromJson(data, &err);
        if (err.error != QJsonParseError::NoError) {
            emitError("NearbyClusters: server response json parse error: " + err.errorString());
            emitStepFinished(*api, step, false, errText);
            return;
        }

        if (!json.isArray()) {
            emitError("NearbyClusters: response is not an array!");
            emitStepFinished(*api, step, false, errText);
            return;
        }

        mResponse = new CashPointResponse;
        mResponse->type = CashPointResponse::CashpointData;

        const QJsonArray arr = json.array();
        const auto end = arr.end();
        for (auto it = arr.begin(); it != end; it++) {
            const QJsonValue &val = *it;
            if (!val.isObject()) {
                qDebug() << "expected json object in array";
                continue;
            }
            const QJsonObject obj = val.toObject();
            if (!obj["size"].isUndefined()) {
                mResponse->addClusterData(obj);
            } else {
                const int id = obj["id"].toInt();
                if (id > 0) {
                    mResponse->addCashPointData(obj);
                    mResponse->addVisiableCashpoint(id);
                }
            }
        }

        emitResponseReady(true);
        emitStepFinished(*api, step, true, trUtf8("List of nearby clusters received"));
    });
}
