#include "icoimageprovider.h"

#include <QtSvg/QSvgRenderer>
#include <QtGui/QPainter>
#include <QDebug>

IcoImageProvider::IcoImageProvider()
    : QQuickImageProvider(QQmlImageProviderBase::Image)
{ }

IcoImageProvider::~IcoImageProvider()
{
    const auto end = mRenderers.begin();
    for (auto it = mRenderers.begin(); it != end; it++) {
        QSvgRenderer *renderer = it.value();
        delete renderer;
    }
}

QImage IcoImageProvider::requestImage(const QString &id, QSize *size, const QSize &requestedSize)
{
    auto it = mRenderers.find(id);
    if (it == mRenderers.end()) {
        if (size) {
            *size = QSize();
        }

        int index = id.indexOf(':');
        if (index >= 0) {
            const QString name = id.left(index);
            const QString arg = id.mid(index + 1);

            auto templIt = mImageTemplates.find(name);
            if (templIt == mImageTemplates.end()) {
                qDebug() << "No such image template:" << name;
                return QImage();
            }

            auto &v = templIt.value();
            QByteArray icoData = v.first;
            const QByteArray &placeholder = v.second;
            icoData.replace(placeholder, arg.toUtf8());

            if (loadSvgImage(id, icoData)) {
                it = mRenderers.find(id);
            } else {
                qDebug() << "Cannot bake svg image from template";
                return QImage();
            }
        } else {
            return QImage();
        }
    }

    QSvgRenderer *renderer = it.value();
//    qDebug() << requestedSize;

    QSize imageSize = renderer->defaultSize();
    if (size) {
        if (requestedSize.isValid()) {
            *size = requestedSize;
            imageSize = requestedSize;
        } else {
            *size = imageSize;
        }
    }
    QImage image(imageSize, QImage::Format_ARGB32);
    image.fill(Qt::transparent);

    QPainter painter(&image);
    renderer->render(&painter, QRectF(QPointF(), requestedSize));

    return image;
}

bool IcoImageProvider::loadSvgImage(const QString &name, const QByteArray &data)
{
    QSvgRenderer *renderer = new QSvgRenderer;
    if (!renderer->load(data)) {
        delete renderer;
        return false;
    }

    unloadSvgImage(name);

    mRenderers[name] = renderer;
    return true;
}

void IcoImageProvider::loadSvgImageTemplate(const QString &name, const QByteArray &data, const QByteArray &placeholder)
{
    mImageTemplates.insert(name, { data, placeholder });
}

bool IcoImageProvider::unloadSvgImage(const QString &name)
{
    auto it = mRenderers.find(name);
    if (it != mRenderers.end()) {
        QSvgRenderer *oldRenderer = it.value();
        mRenderers.erase(it);
        delete oldRenderer;
        return true;
    }
    return false;
}
