import Link from "next/link";

export default function LandingPage() {
  return (
    <div className="page">
      <main className="container hero">
        <h1>Pulse is where your people pulse back.</h1>
        <p>
          Build circles, host groups, and keep your everyday updates flowing. Designed for
          intentional sharing with the people who actually matter.
        </p>
        <div style={{ display: "flex", gap: 12, marginTop: 24, flexWrap: "wrap" }}>
          <Link className="button" href="/register">
            Create account
          </Link>
          <Link className="button ghost" href="/feed">
            Explore feed
          </Link>
        </div>
      </main>

      <section className="container" style={{ paddingBottom: 72 }}>
        <div className="grid two">
          <div className="surface card">
            <h3>Signal-first feed</h3>
            <p>
              See posts from the people you follow and the groups you belong to, with
              privacy that respects your boundaries.
            </p>
          </div>
          <div className="surface card">
            <h3>Groups with intention</h3>
            <p>
              Organize by events, member requests, and curated posts. Every group has its
              own culture.
            </p>
          </div>
          <div className="surface card">
            <h3>Direct chat</h3>
            <p>
              Jump into conversations fast. Chat threads stay anchored to the people or
              groups you care about.
            </p>
          </div>
          <div className="surface card">
            <h3>Privacy controls</h3>
            <p>
              Switch between public, semi-private, and private posts. Share selectively
              without friction.
            </p>
          </div>
        </div>
      </section>
    </div>
  );
}
