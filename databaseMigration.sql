CREATE TABLE VisitedLinks (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  page_rank_score DECIMAL(10,2),
  link VARCHAR(2048) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE UnvisitedLinks (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  link VARCHAR(2048) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE LinkGraph (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  source VARCHAR(2048) NOT NULL,
  destination VARCHAR(2048) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX LinkGraph_Source_Index ON LinkGraph (source);
CREATE INDEX LinkGraph_Destination_Index ON LinkGraph (destination);